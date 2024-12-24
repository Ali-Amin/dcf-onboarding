package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"clever.secure-onboard.com/internal/agent/clients"
	"clever.secure-onboard.com/internal/config"
	"clever.secure-onboard.com/internal/logging"
)

func main() {
	// Load vrrs
	var onboardingServiceURL string
	flag.StringVar(&onboardingServiceURL,
		"onboarding-service-url",
		"./cmd/onboarder/res/config-k8s.json",
		"Location of the onboarding service (e.g., http://180.16.12.5:35000)")
	flag.Parse()

	logger := logging.NewDefaultLogger(config.LoggingInfo{MinLogLevel: "debug"})

	id, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	deviceID := strings.Trim(string(id), "\n")

	onboarder := clients.NewOnboardingServerClient(deviceID, onboardingServiceURL, logger)
	hasTPM := false
	fi, err := os.Stat("/dev/tpm0")
	if err == nil {
		// TPM mounted at default path
		if fi.Mode()&os.ModeDevice != 0 || fi.Mode()&os.ModeSocket != 0 {
			hasTPM = true
		}
	}

	for {
		status, err := onboarder.SendTPMStatus(hasTPM)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		if status == http.StatusServiceUnavailable {
			time.Sleep(5 * time.Second)
			continue
		}

		break
	}

	if !hasTPM {
		logger.Write(slog.LevelInfo, "shutting down dcf agent, no tpm found")
		os.Exit(0)
	}

	challenge, err := onboarder.RequestChallenge()
	if err != nil {
		logger.Error("Failed to get challenge: " + err.Error())
		os.Exit(1)
	}

	// TODO: Use TPM to sign challenge and send signature to onboarding service

	privKey := []byte("77868480DA12295006B2EC15097BCCAF1921316A86E3BE8DC2F0D316E8A3D73B")
	answer := ed25519.Sign(privKey, []byte(challenge))
	answerHEX := hex.EncodeToString(answer)

	passed, err := onboarder.SendChallengeAnswer(answerHEX)
	if err != nil {
		logger.Error("Failed to send challenge answer: " + err.Error())
		os.Exit(1)
	}

	logger.Write(slog.LevelInfo, fmt.Sprintf("Passed: %s", passed))

	if !passed {
		logger.Write(slog.LevelInfo, "shutting down")
		os.Exit(1)
	}

	// TODO: Implement annotators for secureboot, os distro, or geolocation
}

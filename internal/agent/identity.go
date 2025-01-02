package agent

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"clever.secure-onboard.com/internal/agent/clients"
	"clever.secure-onboard.com/pkg/interfaces"
)

type IdentityClaimWorker struct {
	tpmClient        interfaces.TPMClient
	onboardingClient *clients.OnboardingServerClient
	ready            chan bool

	logger interfaces.Logger
}

func NewIdentityClaimWorker(
	tpmClient interfaces.TPMClient,
	onboardingClient *clients.OnboardingServerClient,
	ready chan bool,
	logger interfaces.Logger,
) *IdentityClaimWorker {
	return &IdentityClaimWorker{
		tpmClient:        tpmClient,
		onboardingClient: onboardingClient,
		ready:            ready,
		logger:           logger,
	}
}

func (w *IdentityClaimWorker) Start() bool {
	hasTPM, _ := w.tpmClient.HasTPM()
	if !hasTPM {
		w.logger.Error("No TPM found, shutting down...")
		return false
	}
	w.trySendTPMStatus(hasTPM)
	w.sendIdentityClaim()

	return true
}

func (w *IdentityClaimWorker) trySendTPMStatus(hasTPM bool) {
	exponentialBackoff := 1
	for {
		status, err := w.onboardingClient.SendTPMStatus(hasTPM)
		if err != nil || status == http.StatusServiceUnavailable {
			w.logger.Error(
				fmt.Sprintf(
					"Retrying to reach onboarding service in %d seconds",
					exponentialBackoff,
				),
			)
			// ceiling is one hour
			if exponentialBackoff < 1*60*60 {
				exponentialBackoff *= 2
			}
			time.Sleep(time.Duration(exponentialBackoff) * time.Second)
			continue
		}
		break
	}
}

func (w *IdentityClaimWorker) sendIdentityClaim() {
	// Repeat challenge verification on exponential back off
	// until it passes to allow for retroactive upload of device
	// public key to onboarding service
	passed := false
	exponentialBackoff := 1

	backoff := func() {
		w.logger.Write(
			slog.LevelInfo,
			fmt.Sprintf("retrying in %d seconds", exponentialBackoff),
		)
		time.Sleep(time.Duration(exponentialBackoff) * time.Second)
	}
	for !passed {
		challenge, err := w.onboardingClient.RequestChallenge()
		if err != nil {
			w.logger.Error("failed to get challange: " + err.Error())
			backoff()
			continue
		}

		// If signature failed, something is wrong with TPM and service should be manually looked at again
		signature, err := w.tpmClient.Sign(challenge)
		if err != nil {
			w.logger.Error("failed to sign challenge: " + err.Error())
			os.Exit(1)
		}

		passed, err = w.onboardingClient.SendChallengeAnswer(signature)
		if err != nil {
			w.logger.Error("Failed to send challenge answer: " + err.Error())
			backoff()
			continue
		}

		w.logger.Write(slog.LevelInfo, fmt.Sprintf("Passed: %t", passed))
		if !passed {
			backoff()

			// ceiling is one hour
			if exponentialBackoff < 1*60*60 {
				exponentialBackoff *= 2
			}
			continue
		}
		passed = true
	}
}

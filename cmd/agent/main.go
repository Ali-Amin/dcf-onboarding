package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"clever.secure-onboard.com/internal/agent"
	"clever.secure-onboard.com/internal/agent/clients"
	"clever.secure-onboard.com/internal/annotators"
	"clever.secure-onboard.com/internal/bootstrap"
	"clever.secure-onboard.com/internal/config"
	"clever.secure-onboard.com/internal/logging"
	alvariumcfg "github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"

	alvariumpkg "github.com/project-alvarium/alvarium-sdk-go/pkg"
)

func main() {
	// Load vars
	var onboardingServiceURL string
	flag.StringVar(&onboardingServiceURL,
		"onboarding-service-url",
		"http://0.0.0.0:3010",
		"Location of the onboarding service (e.g., http://180.16.12.5:35000)")

	var alvariumCFGPath string
	flag.StringVar(&alvariumCFGPath,
		"cfg",
		"./cmd/agent/res/config.json",
		"Path of the alvarium SDK config",
	)
	flag.Parse()

	logger := logging.NewDefaultLogger(config.LoggingInfo{MinLogLevel: "debug"})

	id, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	deviceID := strings.Trim(string(id), "\n")

	onboarder := clients.NewOnboardingServerClient(deviceID, onboardingServiceURL, logger)
	tpmClient := clients.NewTPMClient()

	ready := make(chan bool)
	identityWorker := agent.NewIdentityClaimWorker(tpmClient, onboarder, ready, logger)

	var alvariumCFG alvariumcfg.SdkInfo
	err = config.NewJsonReader().Read(alvariumCFGPath, &alvariumCFG)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	alvariumSDK := alvariumpkg.NewSdk([]interfaces.Annotator{
		annotators.NewSecureBootAnnotator(alvariumCFG),
	}, alvariumCFG, logger)

	annotatorWorker := agent.NewAnnotatorWorker(ready, deviceID, alvariumSDK, logger)

	ctx, cancel := context.WithCancel(context.Background())
	bootstrap.Run(
		ctx,
		cancel,
		nil,
		[]bootstrap.BootstrapHandler{
			identityWorker.Start,
			alvariumSDK.BootstrapHandler,
			annotatorWorker.Start,
		})
}

package main

import (
	"context"
	"flag"
	"os"
	"sync"

	"clever.secure-onboard.com/internal/agent"
	"clever.secure-onboard.com/internal/agent/clients"
	"clever.secure-onboard.com/internal/annotators"
	"clever.secure-onboard.com/internal/config"
	"clever.secure-onboard.com/internal/logging"
	"clever.secure-onboard.com/pkg/factories"
	"github.com/project-alvarium/alvarium-sdk-go/pkg"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
)

func main() {
	// Load vars
	var onboardingServiceURL string
	flag.StringVar(&onboardingServiceURL,
		"onboarding-service-url",
		"http://0.0.0.0:3010",
		"Location of the onboarding service (e.g., http://180.16.12.5:35000)")

	var cfg string
	flag.StringVar(&cfg,
		"cfg",
		"./cmd/agent/res/config.json",
		"Path of the DCF Agent config",
	)
	flag.Parse()

	logger := logging.NewDefaultLogger(config.LoggingInfo{MinLogLevel: "debug"})

	// This would give the same machine ID for VMs running on the same host
	// which would work in theory since we are measuring confidence in the host itself,
	// but in practice we would be duplicating annotations which would sway confidence
	// scores one way or the other
	//	id, err := os.ReadFile("/etc/machine-id")
	//	if err != nil {
	//		logger.Error(err.Error())
	//		os.Exit(1)
	//	}
	// deviceID := strings.Trim(string(id), "\n")

	deviceID, _ := os.Hostname()

	onboarder := clients.NewOnboardingServerClient(deviceID, onboardingServiceURL, logger)

	var agentCFG config.DCFAgentConfig
	err := config.NewJsonReader().Read(cfg, &agentCFG)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	tpmClient, err := factories.NewTPMClient(agentCFG.TPMInfo, logger)
	if err != nil {
		logger.Error("Failed to initialize tpm client: " + err.Error())
		os.Exit(1)
	}

	ready := make(chan bool)
	identityWorker := agent.NewIdentityClaimWorker(tpmClient, onboarder, ready, logger)

	alvariumSDK := pkg.NewSdk(
		[]interfaces.Annotator{
			annotators.NewSecureBootAnnotator(agentCFG.AlvariumSDKInfo),
		},
		agentCFG.AlvariumSDKInfo,
		logger,
	)

	annotatorWorker := agent.NewAnnotatorWorker(ready, deviceID, alvariumSDK, logger)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		ctx, _ := context.WithCancel(context.Background())
		alvariumSDK.BootstrapHandler(ctx, wg)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		identityWorker.Start()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		annotatorWorker.Start()
	}()

	wg.Wait()
}

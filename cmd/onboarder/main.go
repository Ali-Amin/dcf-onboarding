package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"clever.secure-onboard.com/internal/agent"
	"clever.secure-onboard.com/internal/annotators"
	"clever.secure-onboard.com/internal/bootstrap"
	"clever.secure-onboard.com/internal/config"
	"clever.secure-onboard.com/internal/logging"
	onboarding "clever.secure-onboard.com/internal/onboarding/server"

	"clever.secure-onboard.com/pkg/contracts"
	"clever.secure-onboard.com/pkg/factories"
	alvarium "github.com/project-alvarium/alvarium-sdk-go/pkg"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
)

func main() {
	// Load config
	var configPath string
	flag.StringVar(&configPath,
		"cfg",
		"./cmd/onboarder/res/config-k8s.json",
		"Path to JSON configuration file.")
	flag.Parse()

	fileFormat := factories.GetFileExtension(configPath)
	reader, err := factories.NewReader(fileFormat)
	if err != nil {
		slog.Log(context.Background(), slog.LevelError, err.Error())
		os.Exit(1)
	}

	cfg := config.OnboardingServiceConfig{}
	err = reader.Read(configPath, &cfg)
	if err != nil {
		slog.Log(context.Background(), slog.LevelError, err.Error())
		os.Exit(1)
	}

	logger := logging.NewDefaultLogger(cfg.Logging)
	logger.Write(slog.LevelDebug, "config loaded successfully")
	logger.Write(slog.LevelDebug, cfg.AsString())

	annotators := []interfaces.Annotator{annotators.NewDeviceIdentityAnnotator(cfg.AlvariumSDK)}
	alvariumSDK := alvarium.NewSdk(annotators, cfg.AlvariumSDK, logger)

	nodeDiscoverer, err := factories.NewNodeDiscoverer(cfg.NodeDiscoverer, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	identityVerifier := factories.NewDeviceIdentityVerifier(alvariumSDK, logger)

	nodeDiscoverer.OnNewNode(func(node contracts.Node) {
		agent.RemoteInstall(cfg.Daemon, []string{node.IP}, logger)
		logger.Write(slog.LevelDebug, fmt.Sprintf("Found node: %s", node))
	})

	auth, err := factories.NewAuthenticator(cfg.OnboardingServer.Auth, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	server := onboarding.NewOnboardingServer(cfg.OnboardingServer, identityVerifier, auth, logger)

	ctx, cancel := context.WithCancel(context.Background())
	bootstrap.Run(ctx, cancel, &cfg, []bootstrap.BootstrapHandler{
		alvariumSDK.BootstrapHandler,
		server.Bootstrap,
		nodeDiscoverer.Bootstrap,
	})
}

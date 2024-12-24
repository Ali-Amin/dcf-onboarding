package onboarding

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"clever.secure-onboard.com/internal/config"
	"clever.secure-onboard.com/pkg/interfaces"
)

type OnboardingServer struct {
	cfg              config.ServerInfo
	logger           interfaces.Logger
	identityVerifier interfaces.DeviceIdentityVerifier
}

func NewOnboardingServer(
	cfg config.ServerInfo,
	identityVerifier interfaces.DeviceIdentityVerifier,
	logger interfaces.Logger,
) *OnboardingServer {
	return &OnboardingServer{
		cfg:              cfg,
		logger:           logger,
		identityVerifier: identityVerifier,
	}
}

func (s *OnboardingServer) Bootstrap(ctx context.Context, wg *sync.WaitGroup) bool {
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)
	s.logger.Write(slog.LevelInfo, "onboarding server running on "+addr)
	r := newRouter(s.identityVerifier, s.logger)
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.ListenAndServe()
		if err != nil {
			s.logger.Error(err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		for {
			<-ctx.Done()
			err := server.Close()
			if err != nil {
				s.logger.Error(err.Error())
			}
			s.logger.Write(slog.LevelInfo, "Shutdown received for onboarding server")
		}
	}()

	return true
}

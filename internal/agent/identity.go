package agent

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"clever.secure-onboard.com/internal/agent/clients"
	"clever.secure-onboard.com/pkg/interfaces"
)

type IdentityClaimWorker struct {
	tpmClient        *clients.TPMClient
	onboardingClient *clients.OnboardingServerClient
	ready            chan bool

	logger interfaces.Logger
}

func NewIdentityClaimWorker(
	tpmClient *clients.TPMClient,
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

func (w *IdentityClaimWorker) Start(ctx context.Context, wg *sync.WaitGroup) bool {
	hasTPM := w.tpmClient.HasTPM()
	if !hasTPM {
		w.logger.Error("No TPM found, shutting down...")
		return false
	}
	w.trySendTPMStatus(ctx, wg, hasTPM)
	w.sendIdentityClaim(ctx, wg)

	return true
}

func (w *IdentityClaimWorker) trySendTPMStatus(
	ctx context.Context,
	wg *sync.WaitGroup,
	hasTPM bool,
) (success chan bool) {
	exponentialBackoff := 1
	for {
		status, err := w.onboardingClient.SendTPMStatus(hasTPM)
		if err != nil || status == http.StatusServiceUnavailable {
			w.logger.Error(
				fmt.Sprintf(
					"Retrying to reach onboarding service in %s seconds",
					exponentialBackoff,
				),
			)
			exponentialBackoff *= 2
			time.Sleep(time.Duration(exponentialBackoff) * time.Second)
			continue
		}
		break
	}
	return nil
}

func (w *IdentityClaimWorker) sendIdentityClaim(ctx context.Context, wg *sync.WaitGroup) {
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

		// TODO: Use TPM to sign challenge and send signature to onboarding service

		privKeyHEX := "77868480DA12295006B2EC15097BCCAF1921316A86E3BE8DC2F0D316E8A3D73B935EA49AF695C3339512F1A814EDBD3BE98651BD2BB521AD4DD75BDB39DD5CA0"
		privKey, err := hex.DecodeString(privKeyHEX)
		if err != nil {
			w.logger.Error("failed to decode private key hex", err.Error())
			return
		}

		answer := ed25519.Sign(privKey, []byte(challenge))
		answerHEX := hex.EncodeToString(answer)

		passed, err = w.onboardingClient.SendChallengeAnswer(answerHEX)
		if err != nil {
			w.logger.Error("Failed to send challenge answer: " + err.Error())
			backoff()
			continue
		}

		w.logger.Write(slog.LevelInfo, fmt.Sprintf("Passed: %t", passed))
		if !passed {
			backoff()
			exponentialBackoff = 5
		} else {
			return
		}
	}
}

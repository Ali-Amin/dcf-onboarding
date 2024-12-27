package agent

import (
	"context"
	"log/slog"
	"sync"

	"clever.secure-onboard.com/pkg/interfaces"
	alvarium "github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
)

type AnnotatorWorker struct {
	ready       chan bool
	deviceID    string
	alvariumSDK alvarium.Sdk
	logger      interfaces.Logger
}

func NewAnnotatorWorker(
	ready chan bool,
	deviceID string,
	alvariumSDK alvarium.Sdk,
	logger interfaces.Logger,
) *AnnotatorWorker {
	return &AnnotatorWorker{
		ready:       ready,
		deviceID:    deviceID,
		alvariumSDK: alvariumSDK,
		logger:      logger,
	}
}

func (w *AnnotatorWorker) Start(ctx context.Context, wg *sync.WaitGroup) bool {
	wg.Add(1)
	defer wg.Done()
	go func() {
		shouldAnnotate := <-w.ready
		if !shouldAnnotate {
			w.logger.Write(slog.LevelInfo, "Received signal to not publishing annotations")
			return
		}
		w.logger.Write(slog.LevelInfo, "Publishing annotations...")
		w.alvariumSDK.Transit(ctx, []byte(w.deviceID))
	}()
	return true
}

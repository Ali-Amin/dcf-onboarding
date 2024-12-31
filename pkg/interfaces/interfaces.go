package interfaces

import (
	"context"
	"log/slog"
	"sync"

	"clever.secure-onboard.com/pkg/contracts"
)

type NodeDiscoverer interface {
	OnNewNode(func(node contracts.Node))
	Bootstrap(ctx context.Context, wg *sync.WaitGroup) bool
}

type DeviceIdentityVerifier interface {
	ReceivePublicKey(deviceID, pubKey string) error
	GenerateChallenge(deviceID string) (challengeToken string, err error)
	VerifyAnswer(deviceID, signature string) (verified bool, err error)
	HandleTPMStatus(deviceID string, hasTPM bool) error
}

type Onboarder interface{}

type TPMClient interface {
	HasTPM() (bool, error)
	Sign(data string) (string, error)
}

type Logger interface {
	// Write facilitates creation and writing of a LogEntry of a specified LogLevel. The client application
	// can also supply a message and a flexible list of additional arguments. These additional arguments are
	// optional. If provided, they should be treated as a key/value pair where the key is of type LogKey.
	//
	// Write flushes the LogEntry to StdOut in JSON format.
	Write(level slog.Level, message string, args ...any)
	// Error facilitates creation and writing of a LogEntry at the Error LogLevel. The client application
	// can also supply a message and a flexible list of additional arguments. These additional arguments are
	// optional. If provided, they should be treated as a key/value pair where the key is of type LogKey.
	//
	// Write flushes the LogEntry to StdErr in JSON format.
	Error(message string, args ...any)
}

type Reader interface {
	Read(filePath string, cfg any) error
}

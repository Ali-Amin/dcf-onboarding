package annotators

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"clever.secure-onboard.com/pkg/contracts"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	sdkContracts "github.com/project-alvarium/alvarium-sdk-go/pkg/contracts"
)

type DeviceIdentityAnnotator struct {
	sdkCFG config.SdkInfo
}

func NewDeviceIdentityAnnotator(sdkCFG config.SdkInfo) *DeviceIdentityAnnotator {
	return &DeviceIdentityAnnotator{sdkCFG: sdkCFG}
}

func (a *DeviceIdentityAnnotator) Do(
	ctx context.Context,
	data []byte,
) (sdkContracts.Annotation, error) {
	key := string(data)          // Device ID of which the annotation belongs to
	hostname, _ := os.Hostname() // Hostname of the host publishing the annotation
	isSatisfied := false
	hasTPM := ctx.Value(contracts.HasTPM)
	if hasTPM != nil {
		hasTPMValue, ok := hasTPM.(bool)
		if ok && hasTPMValue {
			isSatisfied = true
		}
	}
	annotation := sdkContracts.NewAnnotation(
		key,
		a.sdkCFG.Hash.Type,
		hostname,
		sdkContracts.Host,
		"remote-tpm",
		isSatisfied,
	)

	prv, err := os.ReadFile(a.sdkCFG.Signature.PrivateKey.Path)
	if err != nil {
		return sdkContracts.Annotation{}, err
	}

	keyDecoded := make([]byte, hex.DecodedLen(len(prv)))
	hex.Decode(keyDecoded, prv)

	b, _ := json.Marshal(annotation)
	signed := ed25519.Sign(keyDecoded, b)
	annotation.Signature = fmt.Sprintf("%x", signed)
	return annotation, nil
}

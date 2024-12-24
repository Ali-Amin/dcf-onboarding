package annotators

import (
	"context"

	"clever.secure-onboard.com/pkg/contracts"
	"github.com/google/uuid"
	sdkContracts "github.com/project-alvarium/alvarium-sdk-go/pkg/contracts"
)

type DeviceIdentityAnnotator struct{}

func (a *DeviceIdentityAnnotator) Do(
	ctx context.Context,
	data []byte,
) (sdkContracts.Annotation, error) {
	key := uuid.NewString()  // TODO: Use something else?
	host := uuid.NewString() // TODO: Use something else?
	isSatisfied := false
	hasTPM := ctx.Value(contracts.HasTPM)
	if hasTPM != nil {
		hasTPMValue, ok := hasTPM.(bool)
		if ok && hasTPMValue {
			isSatisfied = true
		}
	}
	return sdkContracts.NewAnnotation(
		key,
		sdkContracts.SHA256Hash,
		host,
		sdkContracts.Host,
		"remote-tpm",
		isSatisfied,
	), nil
}

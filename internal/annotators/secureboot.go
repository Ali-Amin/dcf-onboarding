package annotators

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/foxboron/go-uefi/efivarfs"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/contracts"
	sdkContracts "github.com/project-alvarium/alvarium-sdk-go/pkg/contracts"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
)

type SecureBootAnnotator struct {
	sdkCFG config.SdkInfo
}

func NewSecureBootAnnotator(sdkCFG config.SdkInfo) interfaces.Annotator {
	return &SecureBootAnnotator{sdkCFG: sdkCFG}
}

func (a *SecureBootAnnotator) Do(ctx context.Context, data []byte) (contracts.Annotation, error) {
	key := string(data)          // Device ID of which the annotation is for
	hostname, _ := os.Hostname() // Hostname of the host publishing the annotation

	efifs := efivarfs.NewFS().Open()
	isSatisfied, _ := efifs.GetSecureBoot()

	annotation := contracts.NewAnnotation(
		key,
		contracts.SHA256Hash,
		hostname,
		contracts.Host,
		"secure-boot",
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

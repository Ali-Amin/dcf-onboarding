package annotators

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/contracts"
	sdkContracts "github.com/project-alvarium/alvarium-sdk-go/pkg/contracts"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
)

type OSAnnotator struct {
	sdkCFG config.SdkInfo
}

func NewOSAnnotator(sdkCFG config.SdkInfo) interfaces.Annotator {
	return &OSAnnotator{sdkCFG: sdkCFG}
}

func (a *OSAnnotator) Do(ctx context.Context, data []byte) (contracts.Annotation, error) {
	key := string(data)          // Device ID of which the annotation is for
	hostname, _ := os.Hostname() // Hostname of the host publishing the annotation

	distro := readOSRelease()
	whitelist := []string{"Ubuntu 22.04.5 LTS"}
	isSatisfied := slices.Contains(whitelist, distro)

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

func readOSRelease() string {
	releaseFile, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(releaseFile), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Split(line, "=")[1]
		}
	}

	return ""
}

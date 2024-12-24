package factories

import (
	"errors"
	"path"
	"strings"

	"clever.secure-onboard.com/internal/config"
	"clever.secure-onboard.com/internal/onboarding/discovery"
	"clever.secure-onboard.com/internal/onboarding/verifier"
	"clever.secure-onboard.com/pkg/interfaces"
	alvarium "github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
)

func NewNodeDiscoverer(
	cfg config.NodeDiscovererInfo,
	logger interfaces.Logger,
) (interfaces.NodeDiscoverer, error) {
	switch cfg.Type {
	case config.K8s:
		k8sCFG, ok := cfg.Config.(config.K8sNodeDiscoveryConfig)
		if !ok {
			return nil, errors.New("bad k8s node discovery config provided")
		}
		return discovery.NewK8sNodeDiscoverer(k8sCFG, logger)
	default:
		return nil, errors.New("unimplemented node discovery type: " + string(cfg.Type))
	}
}

func NewDeviceIdentityVerifier(
	alvariumSDK alvarium.Sdk,
	logger interfaces.Logger,
) interfaces.DeviceIdentityVerifier {
	return verifier.NewChallengeIdentityVerifier(alvariumSDK)
}

// NewReader returns a type that will hydrate an ApplicationConfig instance from a file.
// Currently only "json" is supported as a value for the readerType parameter. Intention
// is to extend to TOML at some point.
func NewReader(readerType string) (interfaces.Reader, error) {
	var reader interfaces.Reader
	if readerType == "json" {
		reader = config.NewJsonReader()
	} else {
		return reader, errors.New("Unsupported readerType value: " + readerType)
	}
	return reader, nil
}

func GetFileExtension(cfgPath string) string {
	tokens := strings.Split(path.Base(cfgPath), ".")
	if len(tokens) == 2 {
		return tokens[1]
	}
	return tokens[0]
}

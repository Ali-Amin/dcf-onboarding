package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"clever.secure-onboard.com/pkg/contracts"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
)

type OnboardingServiceConfig struct {
	NodeDiscoverer   NodeDiscovererInfo `json:"discovery,omitempty"`
	Daemon           DaemonInfo         `json:"daemon,omitempty"`
	OnboardingServer ServerInfo         `json:"server,omitempty"`
	AlvariumSDK      config.SdkInfo     `json:"sdk,omitempty"`
	Logging          LoggingInfo        `json:"logging,omitempty"`
}

type DCFAgentConfig struct {
	AlvariumSDKInfo config.SdkInfo `json:"alvarium,omitempty"`
	TPMInfo         TPMInfo        `json:"tpm,omitempty"`
}

type NodeDiscovererInfo struct {
	Type   contracts.NodeDiscovererType `json:"type,omitempty"`
	Config interface{}                  `json:"config,omitempty"`
}

// Node discovery info for onboarding
type K8sNodeDiscoveryConfig struct {
	InCluster      bool   `json:"inCluster,omitempty"`
	KubeconfigPath string `json:"kubeconfigPath,omitempty"`
}

func (d *NodeDiscovererInfo) UnmarshalJSON(data []byte) error {
	alias := struct {
		Type contracts.NodeDiscovererType `json:"type,omitempty"`
	}{}

	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}

	switch alias.Type {
	case contracts.K8s:
		k8sAlias := struct {
			Type   contracts.NodeDiscovererType `json:"type,omitempty"`
			Config K8sNodeDiscoveryConfig       `json:"config,omitempty"`
		}{}
		err := json.Unmarshal(data, &k8sAlias)
		if err != nil {
			return err
		}
		d.Type = k8sAlias.Type
		d.Config = k8sAlias.Config
		return nil
	default:
		return fmt.Errorf("unknown node discovery type: %s", alias.Type)
	}
}

type DaemonInfo struct {
	PlaybookPath    string `json:"playbook,omitempty"`
	OnboardingURL   string `json:"onboardingUrl,omitempty"`
	BinaryPath      string `json:"binaryPath,omitempty"`
	SystemdUnitPath string `json:"systemdUnitPath,omitempty"`
	ConfigPath      string `json:"configPath,omitempty"`
	PrivKeyPath     string `json:"privKeyPath,omitempty"`
}

func (c *OnboardingServiceConfig) AsString() string {
	return fmt.Sprintf("%+v", c)
}

type LoggingInfo struct {
	MinLogLevel string `json:"minLogLevel,omitempty"`
}

type ServerInfo struct {
	Protocol string `json:"protocol,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
}

type TPMInfo struct {
	Type   contracts.TPMType `json:"type,omitempty"`
	Config interface{}       `json:"config,omitempty"`
}

type TCPCLIConfig struct {
	PublicKey string `json:"public,omitempty"`
}

type TCPTPMConfig struct {
	Port       string `json:"port,omitempty"`
	Host       string `json:"host,omitempty"`
	PrimaryKey string `json:"primary,omitempty"`
	PublicKey  string `json:"public,omitempty"`
	PrivateKey string `json:"private,omitempty"`
}

func (t *TPMInfo) UnmarshalJSON(data []byte) error {
	alias := &struct {
		Type contracts.TPMType `json:"type,omitempty"`
	}{}

	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}

	switch alias.Type {
	case contracts.CLI:
		alias := &struct {
			Type   contracts.TPMType `json:"type,omitempty"`
			Config TCPCLIConfig      `json:"config,omitempty"`
		}{}

		err := json.Unmarshal(data, &alias)
		if err != nil {
			return err
		}
		t.Type = alias.Type
		t.Config = alias.Config
		return nil
	case contracts.TCP:
		alias := &struct {
			Type   contracts.TPMType `json:"type,omitempty"`
			Config TCPTPMConfig      `json:"config,omitempty"`
		}{}

		err := json.Unmarshal(data, &alias)
		if err != nil {
			return err
		}
		t.Type = alias.Type
		t.Config = alias.Config
		return nil
	default:
		return errors.New("unknown tpm provider type: " + string(alias.Type))

	}
}

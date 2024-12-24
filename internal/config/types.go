package config

import (
	"encoding/json"
	"fmt"

	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
)

type OnboardingServiceConfig struct {
	NodeDiscoverer   NodeDiscovererInfo `json:"discovery,omitempty"`
	Daemon           DaemonInfo         `json:"daemon,omitempty"`
	OnboardingServer ServerInfo         `json:"server,omitempty"`
	AlvariumSDK      config.SdkInfo     `json:"sdk,omitempty"`
	Logging          LoggingInfo        `json:"logging,omitempty"`
}

type NodeDiscovererType string

const (
	K8s NodeDiscovererType = "k8s"
)

type NodeDiscovererInfo struct {
	Type   NodeDiscovererType `json:"type,omitempty"`
	Config interface{}        `json:"config,omitempty"`
}

// Node discovery info for onboarding
type K8sNodeDiscoveryConfig struct {
	InCluster      bool   `json:"inCluster,omitempty"`
	KubeconfigPath string `json:"kubeconfigPath,omitempty"`
}

func (d *NodeDiscovererInfo) UnmarshalJSON(data []byte) error {
	alias := struct {
		Type NodeDiscovererType `json:"type,omitempty"`
	}{}

	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}

	switch alias.Type {
	case K8s:
		k8sAlias := struct {
			Type   NodeDiscovererType     `json:"type,omitempty"`
			Config K8sNodeDiscoveryConfig `json:"config,omitempty"`
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

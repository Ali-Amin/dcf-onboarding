package contracts

import corev1 "k8s.io/api/core/v1"

type AnnotatorContext string

const HasTPM AnnotatorContext = "hasTPM"

type Node struct {
	ID string
	IP string
}

func NewNodeFromK8sNode(node corev1.Node) Node {
	var internalIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			internalIP = addr.Address
		}
	}
	return Node{ID: string(node.UID), IP: internalIP}
}

type AnsibleVars string

const (
	AgentPath            AnsibleVars = "agent_binary_path"
	AgentSystemdUnitPath AnsibleVars = "systemd_unit_path"
	OnboarderURL         AnsibleVars = "onboarder_url"
	CFGPath              AnsibleVars = "dcf_config"
	PrivKeyPath          AnsibleVars = "private_key"
)

type TPMType string

const (
	TCP TPMType = "tcp"
	CLI TPMType = "cli"
)

type TPMConstant string

const NoHandle TPMConstant = "no-handle"

type NodeDiscovererType string

const (
	K8s NodeDiscovererType = "k8s"
)

type AuthenticatorType string

const FixedBasic AuthenticatorType = "basic-fixed"

const (
	TrustedActorUsername = "ONBOARDER_USERNAME"
	TrustedActorPassowrd = "ONBOARDER_PASSWORD"
)

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
	AgentPath            AnsibleVars = "agent_binary_path" //"./cmd/agent/dcfagent"
	AgentSystemdUnitPath AnsibleVars = "systemd_unit_path" //"./scripts/systemd2/dcfagent.service.j2"
)

package discovery

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"clever.secure-onboard.com/internal/config"
	"clever.secure-onboard.com/pkg/contracts"
	"clever.secure-onboard.com/pkg/interfaces"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sNodeDiscoverer struct {
	client *kubernetes.Clientset
	logger interfaces.Logger
	cfg    config.K8sNodeDiscoveryConfig

	onNewNode func(node contracts.Node)
}

func NewK8sNodeDiscoverer(
	cfg config.K8sNodeDiscoveryConfig,
	logger interfaces.Logger,
) (*K8sNodeDiscoverer, error) {
	var k8sCFG *rest.Config
	var err error
	if cfg.InCluster {
		k8sCFG, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		k8sCFG, err = clientcmd.BuildConfigFromFlags("", cfg.KubeconfigPath)
		if err != nil {
			return nil, err
		}

	}
	client, err := kubernetes.NewForConfig(k8sCFG)
	if err != nil {
		return nil, err
	}
	return &K8sNodeDiscoverer{client: client, logger: logger}, nil
}

func (d *K8sNodeDiscoverer) Bootstrap(ctx context.Context, wg *sync.WaitGroup) bool {
	nodes, err := d.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		d.logger.Error("failed to list nodes: " + err.Error())
		return false
	}

	for _, node := range nodes.Items {
		if d.onNewNode != nil {
			d.logger.Write(slog.LevelInfo, fmt.Sprintf("Discovered node %s", node.Status.Addresses))
			d.onNewNode(contracts.NewNodeFromK8sNode(node))
		}
	}

	w, err := d.client.CoreV1().
		Nodes().
		Watch(context.Background(), metav1.ListOptions{ResourceVersion: nodes.ResourceVersion})
	if err != nil {
		d.logger.Error("failed to watch nodes: " + err.Error())
		return false
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			event := <-w.ResultChan()
			switch event.Type {
			case watch.Added:
				switch e := event.Object.(type) {
				case *corev1.Node:
					d.logger.Write(slog.LevelInfo, fmt.Sprintf("Discovered new node %s", e.Status.Addresses))
					go d.onNewNode(contracts.NewNodeFromK8sNode(*e))
				}
			}
		}
	}()

	wg.Add(1)
	go func() { // TODO: Find an actual fix
		time.Sleep(20 * time.Minute)
		w, _ = d.client.CoreV1().
			Nodes().
			Watch(context.Background(), metav1.ListOptions{ResourceVersion: nodes.ResourceVersion})
	}()

	wg.Add(1)
	go func() { // Graceful shutdown
		defer wg.Done()

		<-ctx.Done()
		w.Stop()
		d.logger.Write(slog.LevelInfo, "shutdown received")
	}()
	return true
}

func (d *K8sNodeDiscoverer) OnNewNode(callback func(node contracts.Node)) {
	d.onNewNode = callback
}

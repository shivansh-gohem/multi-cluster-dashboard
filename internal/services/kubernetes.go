package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"multi-cluster-dashboard/internal/models"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesService manages connections to multiple Kubernetes clusters
type KubernetesService struct {
	clients     map[string]*kubernetes.Clientset
	configs     []models.ClusterConfig
	mu          sync.RWMutex
	configPath  string
}

// NewKubernetesService creates a new Kubernetes service
func NewKubernetesService(configPath string) (*KubernetesService, error) {
	svc := &KubernetesService{
		clients:    make(map[string]*kubernetes.Clientset),
		configPath: configPath,
	}

	if err := svc.loadConfigs(); err != nil {
		return nil, fmt.Errorf("failed to load cluster configs: %w", err)
	}

	if err := svc.initializeClients(); err != nil {
		return nil, fmt.Errorf("failed to initialize clients: %w", err)
	}

	return svc, nil
}

// loadConfigs loads cluster configurations from YAML file
func (s *KubernetesService) loadConfigs() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return err
	}

	var config models.ClustersConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	s.configs = config.Clusters
	return nil
}

// initializeClients creates Kubernetes clients for each configured cluster
func (s *KubernetesService) initializeClients() error {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}

	for _, clusterCfg := range s.configs {
		if !clusterCfg.Enabled {
			continue
		}

		config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{CurrentContext: clusterCfg.Context},
		).ClientConfig()
		if err != nil {
			fmt.Printf("Warning: Failed to create config for cluster %s: %v\n", clusterCfg.Name, err)
			continue
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Printf("Warning: Failed to create clientset for cluster %s: %v\n", clusterCfg.Name, err)
			continue
		}

		s.mu.Lock()
		s.clients[clusterCfg.Name] = clientset
		s.mu.Unlock()
	}

	return nil
}

// GetConfigs returns all cluster configurations
func (s *KubernetesService) GetConfigs() []models.ClusterConfig {
	return s.configs
}

// GetClient returns the Kubernetes client for a specific cluster
func (s *KubernetesService) GetClient(clusterName string) (*kubernetes.Clientset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, ok := s.clients[clusterName]
	return client, ok
}

// CheckConnectivity verifies if a cluster is reachable
func (s *KubernetesService) CheckConnectivity(ctx context.Context, clusterName string) bool {
	client, ok := s.GetClient(clusterName)
	if !ok {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	return err == nil
}

// GetNodes returns all nodes for a cluster
func (s *KubernetesService) GetNodes(ctx context.Context, clusterName string) ([]models.Node, error) {
	client, ok := s.GetClient(clusterName)
	if !ok {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}

	nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]models.Node, 0, len(nodeList.Items))
	for _, n := range nodeList.Items {
		node := models.Node{
			Name:    n.Name,
			Status:  getNodeStatus(n),
			Roles:   getNodeRoles(n),
			Age:     formatAge(n.CreationTimestamp.Time),
			Version: n.Status.NodeInfo.KubeletVersion,
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetPods returns all pods for a cluster (optionally filtered by namespace)
func (s *KubernetesService) GetPods(ctx context.Context, clusterName, namespace string) ([]models.Pod, error) {
	client, ok := s.GetClient(clusterName)
	if !ok {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}

	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods := make([]models.Pod, 0, len(podList.Items))
	for _, p := range podList.Items {
		restarts := 0
		for _, cs := range p.Status.ContainerStatuses {
			restarts += int(cs.RestartCount)
		}

		pod := models.Pod{
			Name:      p.Name,
			Namespace: p.Namespace,
			Status:    string(p.Status.Phase),
			Restarts:  restarts,
			Age:       formatAge(p.CreationTimestamp.Time),
			Node:      p.Spec.NodeName,
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

// GetPodSummary returns pod counts by status
func (s *KubernetesService) GetPodSummary(ctx context.Context, clusterName string) (running, pending, failed, total int, err error) {
	pods, err := s.GetPods(ctx, clusterName, "")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	for _, pod := range pods {
		switch pod.Status {
		case "Running":
			running++
		case "Pending":
			pending++
		case "Failed":
			failed++
		}
		total++
	}

	return running, pending, failed, total, nil
}

// GetNodeCount returns the number of nodes in a cluster
func (s *KubernetesService) GetNodeCount(ctx context.Context, clusterName string) (int, error) {
	nodes, err := s.GetNodes(ctx, clusterName)
	if err != nil {
		return 0, err
	}
	return len(nodes), nil
}

// Helper functions
func getNodeStatus(node corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

func getNodeRoles(node corev1.Node) []string {
	roles := []string{}
	for label := range node.Labels {
		if strings.HasPrefix(label, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
			if role != "" {
				roles = append(roles, role)
			}
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	return roles
}

func formatAge(t time.Time) string {
	duration := time.Since(t)
	if duration.Hours() > 24 {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	if duration.Hours() >= 1 {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	if duration.Minutes() >= 1 {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	return fmt.Sprintf("%ds", int(duration.Seconds()))
}

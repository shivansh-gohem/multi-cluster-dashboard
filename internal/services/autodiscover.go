package services

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/fsnotify/fsnotify"
    "gopkg.in/yaml.v3"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
    metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type ClusterInfo struct {
    Name          string
    DisplayName   string
    Clientset     *kubernetes.Clientset
    MetricsClient *metricsclientset.Clientset
    Reachable     bool
}

func (c *ClusterInfo) GetUtilization() (float64, float64) {
    if c.MetricsClient == nil || c.Clientset == nil {
        return 0.0, 0.0
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    nodeMetrics, err := c.MetricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
    if err != nil {
        return 0.0, 0.0
    }
    nodes, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return 0.0, 0.0
    }

    var totalCpuReq, totalCpuCap float64
    var totalMemReq, totalMemCap float64

    for _, nm := range nodeMetrics.Items {
        totalCpuReq += float64(nm.Usage.Cpu().MilliValue())
        totalMemReq += float64(nm.Usage.Memory().Value())
    }

    for _, n := range nodes.Items {
        totalCpuCap += float64(n.Status.Capacity.Cpu().MilliValue())
        totalMemCap += float64(n.Status.Capacity.Memory().Value())
    }

    cpuPercent := 0.0
    if totalCpuCap > 0 {
        cpuPercent = (totalCpuReq / totalCpuCap) * 100
    }

    memPercent := 0.0
    if totalMemCap > 0 {
        memPercent = (totalMemReq / totalMemCap) * 100
    }

    return cpuPercent, memPercent
}

type ClusterRegistry struct {
    mu         sync.RWMutex
    configPath string
    clusters   map[string]*ClusterInfo
    onChange   func(clusters map[string]*ClusterInfo)
}

type yamlConfig struct {
    Clusters []struct {
        Name        string `yaml:"name"`
        DisplayName string `yaml:"displayName"`
        Context     string `yaml:"context"`
        Enabled     bool   `yaml:"enabled"`
    } `yaml:"clusters"`
}

func NewClusterRegistry(configPath string, onChange func(map[string]*ClusterInfo)) *ClusterRegistry {
    return &ClusterRegistry{
        configPath: configPath,
        clusters:   make(map[string]*ClusterInfo),
        onChange:   onChange,
    }
}

func (r *ClusterRegistry) Start(ctx context.Context) error {
    kubeconfigPath := getKubeconfigPath()

    r.refresh(kubeconfigPath)

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return fmt.Errorf("failed to create watcher: %w", err)
    }

    if err := watcher.Add(filepath.Dir(kubeconfigPath)); err != nil {
        log.Printf("Warning: failed to watch kubeconfig dir: %v", err)
    }

    if r.configPath != "" {
        if err := watcher.Add(filepath.Dir(r.configPath)); err != nil {
            log.Printf("Warning: failed to watch config directory: %v", err)
        }
    }

    go func() {
        defer watcher.Close()
        var debounce <-chan time.Time

        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                isKubeconfig := filepath.Base(event.Name) == filepath.Base(kubeconfigPath)
                isClusterConfig := filepath.Base(event.Name) == filepath.Base(r.configPath)

                if isKubeconfig || isClusterConfig {
                    debounce = time.After(1500 * time.Millisecond)
                }

            case <-debounce:
                log.Println("🔄 configuration changed — re-discovering clusters...")
                r.refresh(kubeconfigPath)

            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Printf("Watcher error: %v", err)

            case <-ctx.Done():
                return
            }
        }
    }()

    return nil
}

func (r *ClusterRegistry) GetAll() map[string]*ClusterInfo {
    r.mu.RLock()
    defer r.mu.RUnlock()

    snapshot := make(map[string]*ClusterInfo, len(r.clusters))
    for k, v := range r.clusters {
        snapshot[k] = v
    }
    return snapshot
}

func (r *ClusterRegistry) refresh(kubeconfigPath string) {
    // YAML is optional: only used for display name overrides, never blocks discovery
    displayOverrides := make(map[string]string)
    if r.configPath != "" {
        if yamlData, err := os.ReadFile(r.configPath); err == nil {
            var cfg yamlConfig
            if err := yaml.Unmarshal(yamlData, &cfg); err == nil {
                for _, c := range cfg.Clusters {
                    if c.DisplayName != "" {
                        displayOverrides[c.Context] = c.DisplayName
                    }
                }
            }
        }
    }

    contexts, err := listContextsFromKubeconfig(kubeconfigPath)
    if err != nil {
        log.Printf("Failed to read kubeconfig: %v", err)
        return
    }

    newClusters := make(map[string]*ClusterInfo)

    for _, contextName := range contexts {
        displayName := friendlyName(contextName)
        if override, ok := displayOverrides[contextName]; ok {
            displayName = override
        }

        clientset, metricsClient, err := buildClientset(kubeconfigPath, contextName)
        if err != nil {
            log.Printf("⚠️  Could not build client for %s: %v", contextName, err)
            newClusters[contextName] = &ClusterInfo{
                Name:        contextName,
                DisplayName: displayName,
                Reachable:   false,
            }
            continue
        }

        reachable := pingCluster(clientset)

        newClusters[contextName] = &ClusterInfo{
            Name:          contextName,
            DisplayName:   displayName,
            Clientset:     clientset,
            MetricsClient: metricsClient,
            Reachable:     reachable,
        }

        if reachable {
            log.Printf("✅ Cluster online: %s", contextName)
            // Auto-install metrics-server if missing
            go ensureMetricsServer(contextName, clientset)
        } else {
            log.Printf("💤 Cluster offline: %s", contextName)
        }
    }

    r.mu.Lock()
    r.clusters = newClusters
    r.mu.Unlock()

    if r.onChange != nil {
        r.onChange(newClusters)
    }
}

func getKubeconfigPath() string {
    if env := os.Getenv("KUBECONFIG"); env != "" {
        return env
    }
    return filepath.Join(homedir.HomeDir(), ".kube", "config")
}

func listContextsFromKubeconfig(kubeconfigPath string) ([]string, error) {
    loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
    config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
        loadingRules, &clientcmd.ConfigOverrides{},
    ).RawConfig()
    if err != nil {
        return nil, err
    }

    var names []string
    for name := range config.Contexts {
        names = append(names, name)
    }
    return names, nil
}

func buildClientset(kubeconfigPath, contextName string) (*kubernetes.Clientset, *metricsclientset.Clientset, error) {
    loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
    overrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
    restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
        loadingRules, overrides,
    ).ClientConfig()
    if err != nil {
        return nil, nil, err
    }
    restConfig.Timeout = 3 * time.Second
    
    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, nil, err
    }
    
    metricsClient, _ := metricsclientset.NewForConfig(restConfig)
    
    return clientset, metricsClient, nil
}

func pingCluster(clientset *kubernetes.Clientset) bool {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    _, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
    return err == nil
}

func friendlyName(contextName string) string {
    switch {
    case len(contextName) > 5 && contextName[:5] == "kind-":
        return "Kind: " + contextName[5:]
    case contextName == "minikube":
        return "Minikube"
    case contextName == "docker-desktop":
        return "Docker Desktop"
    default:
        return contextName
    }
}

// ===== Auto-Install Metrics Server =====

const metricsServerManifest = "https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml"

func ensureMetricsServer(contextName string, clientset *kubernetes.Clientset) {
    if hasMetricsServer(clientset) {
        return
    }

    log.Printf("📦 metrics-server not found on %s — auto-installing...", contextName)

    // Step 1: Apply the metrics-server manifest
    applyCmd := exec.Command("kubectl", "--context", contextName, "apply", "-f", metricsServerManifest)
    if output, err := applyCmd.CombinedOutput(); err != nil {
        log.Printf("⚠️  Failed to install metrics-server on %s: %v\n%s", contextName, err, string(output))
        return
    }
    log.Printf("✅ metrics-server installed on %s", contextName)

    // Step 2: Patch for Kind/Minikube self-signed certs (--kubelet-insecure-tls)
    if strings.HasPrefix(contextName, "kind-") || contextName == "minikube" {
        patchJSON := `[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]`
        patchCmd := exec.Command("kubectl", "--context", contextName, "-n", "kube-system",
            "patch", "deployment", "metrics-server", "--type=json", "-p", patchJSON)
        if output, err := patchCmd.CombinedOutput(); err != nil {
            log.Printf("⚠️  Failed to patch metrics-server on %s: %v\n%s", contextName, err, string(output))
        } else {
            log.Printf("✅ metrics-server patched for insecure TLS on %s", contextName)
        }
    }
}

func hasMetricsServer(clientset *kubernetes.Clientset) bool {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := clientset.AppsV1().Deployments("kube-system").Get(ctx, "metrics-server", metav1.GetOptions{})
    return err == nil
}

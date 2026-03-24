package handlers

import (
	"context"
	"net/http"
	"time"

	"multi-cluster-dashboard/internal/models"
	"multi-cluster-dashboard/internal/services"
	"multi-cluster-dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

// APIHandler handles all API requests
type APIHandler struct {
	k8sService  *services.KubernetesService
	promService *services.PrometheusService
	store       *store.MetricsStore
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(k8s *services.KubernetesService, prom *services.PrometheusService, s *store.MetricsStore) *APIHandler {
	return &APIHandler{
		k8sService:  k8s,
		promService: prom,
		store:       s,
	}
}

// GetClusters returns all clusters with their health status
func (h *APIHandler) GetClusters(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	configs := h.k8sService.GetConfigs()
	clusters := make([]models.Cluster, 0, len(configs))

	for _, cfg := range configs {
		if !cfg.Enabled {
			continue
		}

		cluster := models.Cluster{
			Name:        cfg.Name,
			DisplayName: cfg.DisplayName,
			Context:     cfg.Context,
			Status:      "Unknown",
			LastUpdated: time.Now(),
		}

		// Check connectivity
		if !h.k8sService.CheckConnectivity(ctx, cfg.Name) {
			cluster.Status = "Critical"
			clusters = append(clusters, cluster)
			continue
		}

		// Get node count
		nodeCount, err := h.k8sService.GetNodeCount(ctx, cfg.Name)
		if err == nil {
			cluster.NodeCount = nodeCount
		}

		// Get pod summary
		running, pending, failed, total, err := h.k8sService.GetPodSummary(ctx, cfg.Name)
		if err == nil {
			cluster.PodCount = total
			_ = running // for future use
		}

		// Get metrics from Prometheus
		if h.promService.CheckConnectivity(ctx, cfg.PrometheusURL) {
			if cpu, err := h.promService.GetCPUUsage(ctx, cfg.PrometheusURL); err == nil {
				cluster.CPUUsage = cpu
			}
			if mem, err := h.promService.GetMemoryUsage(ctx, cfg.PrometheusURL); err == nil {
				cluster.MemoryUsage = mem
			}
		}

		// Determine health status
		cluster.Status = determineClusterStatus(cluster.CPUUsage, cluster.MemoryUsage, pending, failed)

		clusters = append(clusters, cluster)
	}

	c.JSON(http.StatusOK, gin.H{
		"clusters": clusters,
		"count":    len(clusters),
	})
}

// GetClusterDetails returns detailed info for a specific cluster
func (h *APIHandler) GetClusterDetails(c *gin.Context) {
	clusterName := c.Param("name")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Find cluster config
	var cfg *models.ClusterConfig
	for _, clusterCfg := range h.k8sService.GetConfigs() {
		if clusterCfg.Name == clusterName {
			cfg = &clusterCfg
			break
		}
	}

	if cfg == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	cluster := models.Cluster{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Context:     cfg.Context,
		Status:      "Unknown",
		LastUpdated: time.Now(),
	}

	// Check connectivity
	if !h.k8sService.CheckConnectivity(ctx, cfg.Name) {
		cluster.Status = "Critical"
		c.JSON(http.StatusOK, cluster)
		return
	}

	// Get detailed metrics
	nodeCount, _ := h.k8sService.GetNodeCount(ctx, cfg.Name)
	cluster.NodeCount = nodeCount

	running, pending, failed, total, _ := h.k8sService.GetPodSummary(ctx, cfg.Name)
	cluster.PodCount = total

	if h.promService.CheckConnectivity(ctx, cfg.PrometheusURL) {
		cluster.CPUUsage, _ = h.promService.GetCPUUsage(ctx, cfg.PrometheusURL)
		cluster.MemoryUsage, _ = h.promService.GetMemoryUsage(ctx, cfg.PrometheusURL)
	}

	cluster.Status = determineClusterStatus(cluster.CPUUsage, cluster.MemoryUsage, pending, failed)

	c.JSON(http.StatusOK, gin.H{
		"cluster":  cluster,
		"running":  running,
		"pending":  pending,
		"failed":   failed,
	})
}

// GetClusterNodes returns nodes for a specific cluster
func (h *APIHandler) GetClusterNodes(c *gin.Context) {
	clusterName := c.Param("name")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	nodes, err := h.k8sService.GetNodes(ctx, clusterName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Find cluster config for Prometheus URL
	var promURL string
	for _, cfg := range h.k8sService.GetConfigs() {
		if cfg.Name == clusterName {
			promURL = cfg.PrometheusURL
			break
		}
	}

	// Enrich with metrics if Prometheus is available
	if promURL != "" && h.promService.CheckConnectivity(ctx, promURL) {
		for i := range nodes {
			nodes[i].CPUUsage, _ = h.promService.GetNodeCPUUsage(ctx, promURL, nodes[i].Name)
			nodes[i].MemUsage, _ = h.promService.GetNodeMemoryUsage(ctx, promURL, nodes[i].Name)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
		"count": len(nodes),
	})
}

// GetClusterPods returns pods for a specific cluster
func (h *APIHandler) GetClusterPods(c *gin.Context) {
	clusterName := c.Param("name")
	namespace := c.Query("namespace")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	pods, err := h.k8sService.GetPods(ctx, clusterName, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pods":  pods,
		"count": len(pods),
	})
}

// GetAlerts returns active alerts
func (h *APIHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.store.GetActiveAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetClusterHistory returns historical metrics for a cluster
func (h *APIHandler) GetClusterHistory(c *gin.Context) {
	clusterName := c.Param("name")

	snapshots, err := h.store.GetSnapshots(clusterName, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": snapshots,
		"count":     len(snapshots),
	})
}

// Helper function to determine cluster health status
func determineClusterStatus(cpuUsage, memUsage float64, pending, failed int) string {
	// Critical conditions
	if cpuUsage > 95 || memUsage > 95 || failed > 0 {
		return "Critical"
	}
	// Warning conditions
	if cpuUsage > 80 || memUsage > 80 || pending > 5 {
		return "Warning"
	}
	return "Healthy"
}

package handlers

import (
	"net/http"

	"multi-cluster-dashboard/internal/services"
	"multi-cluster-dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

// PageHandler handles page rendering
type PageHandler struct {
	k8sService  *services.KubernetesService
	promService *services.PrometheusService
	store       *store.MetricsStore
}

// NewPageHandler creates a new page handler
func NewPageHandler(k8s *services.KubernetesService, prom *services.PrometheusService, s *store.MetricsStore) *PageHandler {
	return &PageHandler{
		k8sService:  k8s,
		promService: prom,
		store:       s,
	}
}

// Dashboard renders the main dashboard page
func (h *PageHandler) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Multi-Cluster Dashboard",
	})
}

// ClusterDetail renders the cluster detail page
func (h *PageHandler) ClusterDetail(c *gin.Context) {
	clusterName := c.Param("name")
	c.HTML(http.StatusOK, "cluster_detail.html", gin.H{
		"title":       clusterName + " - Cluster Details",
		"clusterName": clusterName,
	})
}

// Alerts renders the alerts page
func (h *PageHandler) Alerts(c *gin.Context) {
	c.HTML(http.StatusOK, "alerts.html", gin.H{
		"title": "Alerts",
	})
}

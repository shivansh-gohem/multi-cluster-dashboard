package handlers

import (
	"net/http"

	"multi-cluster-dashboard/internal/services"
	"multi-cluster-dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

type PageHandler struct {
	registry *services.ClusterRegistry
	store    *store.MetricsStore
}

func NewPageHandler(registry *services.ClusterRegistry, s *store.MetricsStore) *PageHandler {
	return &PageHandler{
		registry: registry,
		store:    s,
	}
}

func (h *PageHandler) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Multi-Cluster Dashboard",
	})
}

func (h *PageHandler) ClusterDetail(c *gin.Context) {
	clusterName := c.Param("name")
	c.HTML(http.StatusOK, "cluster_detail.html", gin.H{
		"title":       clusterName + " - Cluster Details",
		"clusterName": clusterName,
	})
}

func (h *PageHandler) Alerts(c *gin.Context) {
	c.HTML(http.StatusOK, "alerts.html", gin.H{
		"title": "Alerts",
	})
}

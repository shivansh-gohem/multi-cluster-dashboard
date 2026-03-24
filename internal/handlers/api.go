package handlers

import (
	"context"
	"net/http"
	"time"

	"multi-cluster-dashboard/internal/services"
	"multi-cluster-dashboard/internal/store"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIHandler struct {
	registry *services.ClusterRegistry
	store    *store.MetricsStore
}

func NewAPIHandler(registry *services.ClusterRegistry, s *store.MetricsStore) *APIHandler {
	return &APIHandler{
		registry: registry,
		store:    s,
	}
}

func (h *APIHandler) GetClusters(c *gin.Context) {
	clusters := h.registry.GetAll()
	var result []gin.H

	for _, cl := range clusters {
		// Only show reachable/online clusters on the dashboard
		if !cl.Reachable {
			continue
		}

		entry := gin.H{
			"name":        cl.Name,
			"displayName": cl.DisplayName,
			"context":     cl.Name,
			"reachable":   cl.Reachable,
			"status":      "Unknown",
			"nodeCount":   0,
			"podCount":    0,
			"cpuUsage":    0.0,
			"memoryUsage": 0.0,
		}

		nodes, _ := cl.Clientset.CoreV1().Nodes().List(
			context.Background(), metav1.ListOptions{},
		)
		pods, _ := cl.Clientset.CoreV1().Pods("").List(
			context.Background(), metav1.ListOptions{},
		)

		running, failed := 0, 0
		for _, p := range pods.Items {
			if p.Status.Phase == "Running" {
				running++
			} else if p.Status.Phase == "Failed" || p.Status.Phase == "CrashLoopBackOff" {
				failed++
			}
		}

		entry["nodeCount"] = len(nodes.Items)
		entry["podCount"] = len(pods.Items)
		entry["runningPods"] = running
		entry["failedPods"] = failed
            
		cpu, mem := cl.GetUtilization()
		entry["cpuUsage"] = cpu
		entry["memoryUsage"] = mem

		if failed > 0 || cpu > 90 || mem > 90 {
			entry["status"] = "Warning"
		} else {
			entry["status"] = "Healthy"
		}

		result = append(result, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"clusters": result,
		"count":    len(result),
	})
}

func (h *APIHandler) GetClusterDetails(c *gin.Context) {
	clusterName := c.Param("name")
	clusters := h.registry.GetAll()
	cl, ok := clusters[clusterName]
	
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	cluster := gin.H{
		"name":        cl.Name,
		"displayName": cl.DisplayName,
		"context":     cl.Name,
		"status":      "Unknown",
		"reachable":   cl.Reachable,
		"nodeCount":   0,
		"podCount":    0,
		"cpuUsage":    0.0,
		"memoryUsage": 0.0,
	}

	if !cl.Reachable {
		cluster["status"] = "Critical"
		c.JSON(http.StatusOK, gin.H{"cluster": cluster})
		return
	}

	nodes, _ := cl.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	pods, _ := cl.Clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})

	running, failed, pending := 0, 0, 0
	for _, p := range pods.Items {
		switch p.Status.Phase {
		case "Running":
			running++
		case "Pending":
			pending++
		case "Failed", "CrashLoopBackOff":
			failed++
		}
	}

	cluster["nodeCount"] = len(nodes.Items)
	cluster["podCount"] = len(pods.Items)

	cpu, mem := cl.GetUtilization()
	cluster["cpuUsage"] = cpu
	cluster["memoryUsage"] = mem

	if failed > 0 || cpu > 90 || mem > 90 {
		cluster["status"] = "Warning"
	} else {
		cluster["status"] = "Healthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"cluster": cluster,
		"running": running,
		"pending": pending,
		"failed":  failed,
	})
}

func (h *APIHandler) GetClusterNodes(c *gin.Context) {
	name := c.Param("name")
	clusters := h.registry.GetAll()
	cl, ok := clusters[name]
	if !ok || !cl.Reachable {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found or unreachable"})
		return
	}

	nodes, err := cl.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []gin.H
	for _, n := range nodes.Items {
		ready := "NotReady"
		for _, cond := range n.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				ready = "Ready"
			}
		}
		
		var roles []string
		for k := range n.Labels {
			if len(k) > 17 && k[:17] == "node-role.kubernetes.io/" {
				roles = append(roles, k[17:])
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "<none>")
		}

		age := time.Since(n.CreationTimestamp.Time).Round(time.Hour)
		ageStr := ""
		if age < 24*time.Hour {
			ageStr = time.Since(n.CreationTimestamp.Time).Truncate(time.Minute).String()
		} else {
			days := int(age.Hours() / 24)
			ageStr = string(rune(days+'0')) + "d"
		}

		result = append(result, gin.H{
			"name":    n.Name,
			"status":  ready,
			"roles":   roles,
			"version": n.Status.NodeInfo.KubeletVersion,
			"age":     ageStr,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"nodes": result,
		"count": len(result),
	})
}

func (h *APIHandler) GetClusterPods(c *gin.Context) {
	name := c.Param("name")
	clusters := h.registry.GetAll()
	cl, ok := clusters[name]
	if !ok || !cl.Reachable {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found or unreachable"})
		return
	}

	pods, err := cl.Clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []gin.H
	for _, p := range pods.Items {
		var restarts int32
		for _, cs := range p.Status.ContainerStatuses {
			restarts += cs.RestartCount
		}

		result = append(result, gin.H{
			"name":      p.Name,
			"namespace": p.Namespace,
			"status":    string(p.Status.Phase),
			"restarts":  restarts,
			"node":      p.Spec.NodeName,
			"age":       time.Since(p.CreationTimestamp.Time).Truncate(time.Minute).String(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"pods":  result,
		"count": len(result),
	})
}

func (h *APIHandler) GetAlerts(c *gin.Context) {
	clusters := h.registry.GetAll()
	var alerts []gin.H

	for _, cl := range clusters {
		if !cl.Reachable {
			alerts = append(alerts, gin.H{
				"cluster":  cl.Name,
				"severity": "Critical",
				"message":  "Cluster is unreachable",
			})
			continue
		}
		pods, err := cl.Clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
		if err == nil {
			failed := 0
			for _, p := range pods.Items {
				if p.Status.Phase == "Failed" || p.Status.Phase == "CrashLoopBackOff" {
					failed++
				}
			}
			if failed > 0 {
				alerts = append(alerts, gin.H{
					"cluster":  cl.Name,
					"severity": "Warning",
					"message":  "Pod(s) in Failed state",
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

func (h *APIHandler) GetClusterHistory(c *gin.Context) {
	clusterName := c.Param("name")

	if h.store != nil {
		snapshots, err := h.store.GetSnapshots(clusterName, 24*time.Hour)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"snapshots": snapshots,
				"count":     len(snapshots),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": []interface{}{},
		"count":     0,
	})
}

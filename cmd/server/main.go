package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"multi-cluster-dashboard/internal/handlers"
	"multi-cluster-dashboard/internal/models"
	"multi-cluster-dashboard/internal/services"
	"multi-cluster-dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

func main() {
	// Configuration
	port := getEnv("PORT", "8080")
	configPath := getEnv("CLUSTER_CONFIG", "k8s-configs/clusters.yaml")
	dbPath := getEnv("DB_PATH", "data/metrics.db")

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize services
	log.Println("Initializing Kubernetes service...")
	k8sService, err := services.NewKubernetesService(configPath)
	if err != nil {
		log.Printf("Warning: Kubernetes service initialization failed: %v", err)
		log.Println("Dashboard will run with limited functionality")
	}

	log.Println("Initializing Prometheus service...")
	promService := services.NewPrometheusService()

	log.Println("Initializing metrics store...")
	metricsStore, err := store.NewMetricsStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize metrics store: %v", err)
	}
	defer metricsStore.Close()

	// Initialize handlers
	apiHandler := handlers.NewAPIHandler(k8sService, promService, metricsStore)
	pageHandler := handlers.NewPageHandler(k8sService, promService, metricsStore)

	// Set up Gin router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("templates/*.html")

	// Serve static files
	router.Static("/static", "./static")

	// Page routes
	router.GET("/", pageHandler.Dashboard)
	router.GET("/cluster/:name", pageHandler.ClusterDetail)
	router.GET("/alerts", pageHandler.Alerts)

	// API routes
	api := router.Group("/api")
	{
		api.GET("/clusters", apiHandler.GetClusters)
		api.GET("/clusters/:name", apiHandler.GetClusterDetails)
		api.GET("/clusters/:name/nodes", apiHandler.GetClusterNodes)
		api.GET("/clusters/:name/pods", apiHandler.GetClusterPods)
		api.GET("/clusters/:name/history", apiHandler.GetClusterHistory)
		api.GET("/alerts", apiHandler.GetAlerts)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Start metrics collection goroutine
	go startMetricsCollector(k8sService, promService, metricsStore)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// startMetricsCollector periodically collects and stores metrics
func startMetricsCollector(k8s *services.KubernetesService, prom *services.PrometheusService, store *store.MetricsStore) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Also cleanup old data periodically
	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ticker.C:
			collectMetrics(k8s, prom, store)
		case <-cleanupTicker.C:
			if err := store.CleanupOldSnapshots(24 * time.Hour); err != nil {
				log.Printf("Failed to cleanup old snapshots: %v", err)
			}
		}
	}
}

// collectMetrics gathers current metrics and stores them
func collectMetrics(k8s *services.KubernetesService, prom *services.PrometheusService, metricsStore *store.MetricsStore) {
	if k8s == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, cfg := range k8s.GetConfigs() {
		if !cfg.Enabled {
			continue
		}

		snapshot := &models.MetricSnapshot{
			Cluster: cfg.Name,
		}

		var failed int

		// Get node count
		if nodeCount, err := k8s.GetNodeCount(ctx, cfg.Name); err == nil {
			snapshot.NodeCount = nodeCount
		}

		// Get pod count and failed pods
		if _, _, failedPods, podCount, err := k8s.GetPodSummary(ctx, cfg.Name); err == nil {
			snapshot.PodCount = podCount
			failed = failedPods
		}

		// Get CPU/Memory from Prometheus
		if prom.CheckConnectivity(ctx, cfg.PrometheusURL) {
			if cpu, err := prom.GetCPUUsage(ctx, cfg.PrometheusURL); err == nil {
				snapshot.CPUUsage = cpu
			}
			if mem, err := prom.GetMemoryUsage(ctx, cfg.PrometheusURL); err == nil {
				snapshot.MemoryUsage = mem
			}
		}

		if err := metricsStore.SaveSnapshot(snapshot); err != nil {
			log.Printf("Failed to save snapshot for %s: %v", cfg.Name, err)
		}

		// Generate alerts based on thresholds
		checkAndCreateAlerts(metricsStore, cfg.Name, snapshot.CPUUsage, snapshot.MemoryUsage, failed)
	}
}

// checkAndCreateAlerts creates alerts when thresholds are breached
func checkAndCreateAlerts(metricsStore *store.MetricsStore, cluster string, cpuUsage, memUsage float64, failedPods int) {
	// CPU alerts
	if cpuUsage > 95 {
		alert := &models.Alert{
			Cluster:  cluster,
			Severity: "Critical",
			Message:  fmt.Sprintf("CPU usage is critically high at %.1f%%", cpuUsage),
		}
		if err := metricsStore.SaveAlert(alert); err != nil {
			log.Printf("Failed to save CPU critical alert: %v", err)
		}
	} else if cpuUsage > 80 {
		alert := &models.Alert{
			Cluster:  cluster,
			Severity: "Warning",
			Message:  fmt.Sprintf("CPU usage is elevated at %.1f%%", cpuUsage),
		}
		if err := metricsStore.SaveAlert(alert); err != nil {
			log.Printf("Failed to save CPU warning alert: %v", err)
		}
	}

	// Memory alerts
	if memUsage > 95 {
		alert := &models.Alert{
			Cluster:  cluster,
			Severity: "Critical",
			Message:  fmt.Sprintf("Memory usage is critically high at %.1f%%", memUsage),
		}
		if err := metricsStore.SaveAlert(alert); err != nil {
			log.Printf("Failed to save memory critical alert: %v", err)
		}
	} else if memUsage > 80 {
		alert := &models.Alert{
			Cluster:  cluster,
			Severity: "Warning",
			Message:  fmt.Sprintf("Memory usage is elevated at %.1f%%", memUsage),
		}
		if err := metricsStore.SaveAlert(alert); err != nil {
			log.Printf("Failed to save memory warning alert: %v", err)
		}
	}

	// Failed pods alert
	if failedPods > 0 {
		alert := &models.Alert{
			Cluster:  cluster,
			Severity: "Warning",
			Message:  fmt.Sprintf("%d pod(s) in failed state", failedPods),
		}
		if err := metricsStore.SaveAlert(alert); err != nil {
			log.Printf("Failed to save pod failure alert: %v", err)
		}
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

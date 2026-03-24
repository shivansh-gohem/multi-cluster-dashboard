package main

import (
	"context"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	port := getEnv("PORT", "8080")
	configPath := getEnv("CLUSTER_CONFIG", "k8s-configs/clusters.yaml")
	dbPath := getEnv("DB_PATH", "data/metrics.db")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	log.Println("Initializing metrics store...")
	metricsStore, err := store.NewMetricsStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize metrics store: %v", err)
	}
	defer metricsStore.Close()

	ctx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	registry := services.NewClusterRegistry(configPath, func(clusters map[string]*services.ClusterInfo) {
		log.Printf("📡 Cluster list updated — %d cluster(s) found", len(clusters))
	})

	if err := registry.Start(ctx); err != nil {
		log.Fatalf("Failed to start cluster registry: %v", err)
	}

	apiHandler := handlers.NewAPIHandler(registry, metricsStore)
	pageHandler := handlers.NewPageHandler(registry, metricsStore)

	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/static", "./static")

	router.GET("/", pageHandler.Dashboard)
	router.GET("/cluster/:name", pageHandler.ClusterDetail)
	router.GET("/alerts", pageHandler.Alerts)

	api := router.Group("/api")
	{
		api.GET("/clusters", apiHandler.GetClusters)
		api.GET("/clusters/:name", apiHandler.GetClusterDetails)
		api.GET("/clusters/:name/nodes", apiHandler.GetClusterNodes)
		api.GET("/clusters/:name/pods", apiHandler.GetClusterPods)
		api.GET("/clusters/:name/history", apiHandler.GetClusterHistory)
		api.GET("/alerts", apiHandler.GetAlerts)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	go startMetricsCollector(registry, metricsStore)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	mainCancel() // Stop the registry watcher

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func startMetricsCollector(registry *services.ClusterRegistry, store *store.MetricsStore) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ticker.C:
			collectMetrics(registry, store)
		case <-cleanupTicker.C:
			if err := store.CleanupOldSnapshots(24 * time.Hour); err != nil {
				log.Printf("Failed to cleanup old snapshots: %v", err)
			}
		}
	}
}

func collectMetrics(registry *services.ClusterRegistry, metricsStore *store.MetricsStore) {
	clusters := registry.GetAll()
	for _, cl := range clusters {
		if !cl.Reachable {
			continue
		}

		snapshot := &models.MetricSnapshot{
			Cluster: cl.Name, // Using the context name for the snapshot
		}

		nodes, err := cl.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err == nil {
			snapshot.NodeCount = len(nodes.Items)
		}

		pods, err := cl.Clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
		if err == nil {
			snapshot.PodCount = len(pods.Items)
		}

		// Fetch CPU/Memory utilizing the metrics api
		cpu, mem := cl.GetUtilization()
		snapshot.CPUUsage = cpu
		snapshot.MemoryUsage = mem

		if err := metricsStore.SaveSnapshot(snapshot); err != nil {
			log.Printf("Failed to save snapshot for %s: %v", cl.Name, err)
		}
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

package store

import (
	"os"
	"testing"
	"time"

	"multi-cluster-dashboard/internal/models"
)

// TestMetricsStoreCreation tests creating a new metrics store
func TestMetricsStoreCreation(t *testing.T) {
	// Use temp file for test database
	tmpFile, err := os.CreateTemp("", "test-metrics-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	store, err := NewMetricsStore(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create metrics store: %v", err)
	}
	defer store.Close()

	if store.db == nil {
		t.Fatal("Database connection is nil")
	}
}

// TestSaveAndGetSnapshot tests saving and retrieving snapshots
func TestSaveAndGetSnapshot(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-metrics-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	store, err := NewMetricsStore(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create metrics store: %v", err)
	}
	defer store.Close()

	// Save a snapshot
	snapshot := &models.MetricSnapshot{
		Cluster:     "test-cluster",
		CPUUsage:    45.5,
		MemoryUsage: 60.2,
		PodCount:    10,
		NodeCount:   3,
	}

	if err := store.SaveSnapshot(snapshot); err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	// Retrieve snapshots
	snapshots, err := store.GetSnapshots("test-cluster", 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to get snapshots: %v", err)
	}

	if len(snapshots) != 1 {
		t.Fatalf("Expected 1 snapshot, got %d", len(snapshots))
	}

	if snapshots[0].Cluster != "test-cluster" {
		t.Errorf("Expected cluster 'test-cluster', got '%s'", snapshots[0].Cluster)
	}
	if snapshots[0].CPUUsage != 45.5 {
		t.Errorf("Expected CPU 45.5, got %f", snapshots[0].CPUUsage)
	}
}

// TestSaveAndGetAlert tests saving and retrieving alerts
func TestSaveAndGetAlert(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-metrics-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	store, err := NewMetricsStore(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create metrics store: %v", err)
	}
	defer store.Close()

	// Save an alert
	alert := &models.Alert{
		Cluster:  "test-cluster",
		Severity: "Warning",
		Message:  "CPU usage high",
	}

	if err := store.SaveAlert(alert); err != nil {
		t.Fatalf("Failed to save alert: %v", err)
	}

	// Retrieve active alerts
	alerts, err := store.GetActiveAlerts()
	if err != nil {
		t.Fatalf("Failed to get alerts: %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].Cluster != "test-cluster" {
		t.Errorf("Expected cluster 'test-cluster', got '%s'", alerts[0].Cluster)
	}
	if alerts[0].Severity != "Warning" {
		t.Errorf("Expected severity 'Warning', got '%s'", alerts[0].Severity)
	}
}

// TestResolveAlert tests resolving an alert
func TestResolveAlert(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-metrics-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	store, err := NewMetricsStore(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create metrics store: %v", err)
	}
	defer store.Close()

	// Save an alert
	alert := &models.Alert{
		Cluster:  "test-cluster",
		Severity: "Critical",
		Message:  "Memory critical",
	}

	if err := store.SaveAlert(alert); err != nil {
		t.Fatalf("Failed to save alert: %v", err)
	}

	// Resolve the alert
	if err := store.ResolveAlert(alert.ID); err != nil {
		t.Fatalf("Failed to resolve alert: %v", err)
	}

	// Active alerts should be empty
	alerts, err := store.GetActiveAlerts()
	if err != nil {
		t.Fatalf("Failed to get alerts: %v", err)
	}

	if len(alerts) != 0 {
		t.Errorf("Expected 0 active alerts after resolve, got %d", len(alerts))
	}
}

package models

import "time"

// Cluster represents a Kubernetes cluster with its health metrics
type Cluster struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"displayName"`
	Status      string  `json:"status"` // Healthy/Warning/Critical
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
	NodeCount   int     `json:"nodeCount"`
	PodCount    int     `json:"podCount"`
	Context     string  `json:"context"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// ClusterConfig represents the configuration for connecting to a cluster
type ClusterConfig struct {
	Name          string `yaml:"name"`
	DisplayName   string `yaml:"displayName"`
	Context       string `yaml:"context"`
	PrometheusURL string `yaml:"prometheusURL"`
	Enabled       bool   `yaml:"enabled"`
}

// ClustersConfig represents the full clusters configuration file
type ClustersConfig struct {
	Clusters []ClusterConfig `yaml:"clusters"`
}

// Node represents a Kubernetes node
type Node struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // Ready/NotReady
	CPUUsage  float64   `json:"cpuUsage"`
	MemUsage  float64   `json:"memUsage"`
	PodCount  int       `json:"podCount"`
	Roles     []string  `json:"roles"`
	Age       string    `json:"age"`
	Version   string    `json:"version"`
}

// Pod represents a Kubernetes pod
type Pod struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Status     string    `json:"status"` // Running/Pending/Failed/Succeeded
	Restarts   int       `json:"restarts"`
	Age        string    `json:"age"`
	Node       string    `json:"node"`
	CPUUsage   float64   `json:"cpuUsage"`
	MemUsage   float64   `json:"memUsage"`
}

// Alert represents a system alert
type Alert struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Cluster   string    `json:"cluster"`
	Severity  string    `json:"severity"` // Warning/Critical
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
}

// MetricSnapshot stores historical metrics for charting
type MetricSnapshot struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Cluster     string    `json:"cluster" gorm:"index"`
	CPUUsage    float64   `json:"cpuUsage"`
	MemoryUsage float64   `json:"memoryUsage"`
	PodCount    int       `json:"podCount"`
	NodeCount   int       `json:"nodeCount"`
	Timestamp   time.Time `json:"timestamp" gorm:"index"`
}

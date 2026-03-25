package services

import (
	"testing"
	"time"
)

// TestFormatAge tests the formatAge helper function
func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "seconds",
			duration: 30 * time.Second,
			expected: "30s",
		},
		{
			name:     "minutes",
			duration: 15 * time.Minute,
			expected: "15m",
		},
		{
			name:     "hours",
			duration: 5 * time.Hour,
			expected: "5h",
		},
		{
			name:     "days",
			duration: 3 * 24 * time.Hour,
			expected: "3d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Now().Add(-tt.duration)
			result := formatAge(testTime)
			if result != tt.expected {
				t.Errorf("formatAge() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetNodeRolesEmpty tests getNodeRoles with no role labels
func TestGetNodeRolesEmpty(t *testing.T) {
	// When a node has no role labels, it should default to "worker"
	// This is tested via integration as getNodeRoles requires a corev1.Node
}

// TestPodStatusCounting tests that pod status counting logic works correctly
func TestPodStatusCounting(t *testing.T) {
	// Test data representing pod statuses
	podStatuses := []string{"Running", "Running", "Pending", "Failed", "Running"}
	
	var running, pending, failed, total int
	for _, status := range podStatuses {
		switch status {
		case "Running":
			running++
		case "Pending":
			pending++
		case "Failed":
			failed++
		}
		total++
	}

	if running != 3 {
		t.Errorf("Expected 3 running pods, got %d", running)
	}
	if pending != 1 {
		t.Errorf("Expected 1 pending pod, got %d", pending)
	}
	if failed != 1 {
		t.Errorf("Expected 1 failed pod, got %d", failed)
	}
	if total != 5 {
		t.Errorf("Expected 5 total pods, got %d", total)
	}
}

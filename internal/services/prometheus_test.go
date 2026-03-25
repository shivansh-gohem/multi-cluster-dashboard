package services

import (
	"testing"
)

// TestPrometheusServiceCreation tests that the service is created correctly
func TestPrometheusServiceCreation(t *testing.T) {
	svc := NewPrometheusService()
	if svc == nil {
		t.Fatal("NewPrometheusService() returned nil")
	}
	if svc.httpClient == nil {
		t.Fatal("PrometheusService httpClient is nil")
	}
}

// TestExtractFirstValueEmpty tests extracting value from empty result
func TestExtractFirstValueEmpty(t *testing.T) {
	svc := NewPrometheusService()
	
	result := &PrometheusResult{
		Status: "success",
	}
	result.Data.ResultType = "vector"
	result.Data.Result = []struct {
		Metric map[string]string `json:"metric"`
		Value  []interface{}     `json:"value"`
	}{}

	value, err := svc.extractFirstValue(result)
	if err != nil {
		t.Errorf("Expected no error for empty result, got %v", err)
	}
	if value != 0 {
		t.Errorf("Expected 0 for empty result, got %f", value)
	}
}

// TestExtractFirstValueValid tests extracting a valid value
func TestExtractFirstValueValid(t *testing.T) {
	svc := NewPrometheusService()
	
	result := &PrometheusResult{
		Status: "success",
	}
	result.Data.ResultType = "vector"
	result.Data.Result = []struct {
		Metric map[string]string `json:"metric"`
		Value  []interface{}     `json:"value"`
	}{
		{
			Metric: map[string]string{},
			Value:  []interface{}{1234567890.0, "42.5"},
		},
	}

	value, err := svc.extractFirstValue(result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if value != 42.5 {
		t.Errorf("Expected 42.5, got %f", value)
	}
}

// TestExtractFirstValueInvalidFormat tests handling of malformed result
func TestExtractFirstValueInvalidFormat(t *testing.T) {
	svc := NewPrometheusService()
	
	result := &PrometheusResult{
		Status: "success",
	}
	result.Data.ResultType = "vector"
	result.Data.Result = []struct {
		Metric map[string]string `json:"metric"`
		Value  []interface{}     `json:"value"`
	}{
		{
			Metric: map[string]string{},
			Value:  []interface{}{1234567890.0}, // Missing second element
		},
	}

	_, err := svc.extractFirstValue(result)
	if err == nil {
		t.Error("Expected error for malformed result, got nil")
	}
}

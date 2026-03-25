package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// PrometheusService handles querying Prometheus for metrics
type PrometheusService struct {
	httpClient *http.Client
}

// PrometheusResult represents a Prometheus query result
type PrometheusResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// NewPrometheusService creates a new Prometheus service
func NewPrometheusService() *PrometheusService {
	return &PrometheusService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Query executes a PromQL query against the given Prometheus endpoint
func (s *PrometheusService) Query(ctx context.Context, prometheusURL, query string) (*PrometheusResult, error) {
	endpoint := fmt.Sprintf("%s/api/v1/query", prometheusURL)
	
	params := url.Values{}
	params.Set("query", query)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result PrometheusResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("Prometheus query failed: %s", result.Status)
	}

	return &result, nil
}

// GetCPUUsage returns the cluster CPU usage percentage
func (s *PrometheusService) GetCPUUsage(ctx context.Context, prometheusURL string) (float64, error) {
	// Query for cluster-wide CPU usage
	query := `100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
	
	result, err := s.Query(ctx, prometheusURL, query)
	if err != nil {
		return 0, err
	}

	return s.extractFirstValue(result)
}

// GetMemoryUsage returns the cluster memory usage percentage
func (s *PrometheusService) GetMemoryUsage(ctx context.Context, prometheusURL string) (float64, error) {
	// Query for cluster-wide memory usage
	query := `100 * (1 - sum(node_memory_MemAvailable_bytes) / sum(node_memory_MemTotal_bytes))`
	
	result, err := s.Query(ctx, prometheusURL, query)
	if err != nil {
		return 0, err
	}

	return s.extractFirstValue(result)
}

// GetPodCPUUsage returns CPU usage for a specific pod
func (s *PrometheusService) GetPodCPUUsage(ctx context.Context, prometheusURL, namespace, podName string) (float64, error) {
	query := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s",pod="%s"}[5m])) * 100`, namespace, podName)
	
	result, err := s.Query(ctx, prometheusURL, query)
	if err != nil {
		return 0, err
	}

	return s.extractFirstValue(result)
}

// GetPodMemoryUsage returns memory usage for a specific pod in bytes
func (s *PrometheusService) GetPodMemoryUsage(ctx context.Context, prometheusURL, namespace, podName string) (float64, error) {
	query := fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace="%s",pod="%s"})`, namespace, podName)
	
	result, err := s.Query(ctx, prometheusURL, query)
	if err != nil {
		return 0, err
	}

	return s.extractFirstValue(result)
}

// GetNodeCPUUsage returns CPU usage for a specific node
func (s *PrometheusService) GetNodeCPUUsage(ctx context.Context, prometheusURL, nodeName string) (float64, error) {
	query := fmt.Sprintf(`100 - (avg(irate(node_cpu_seconds_total{mode="idle",instance=~"%s.*"}[5m])) * 100)`, nodeName)
	
	result, err := s.Query(ctx, prometheusURL, query)
	if err != nil {
		return 0, err
	}

	return s.extractFirstValue(result)
}

// GetNodeMemoryUsage returns memory usage for a specific node
func (s *PrometheusService) GetNodeMemoryUsage(ctx context.Context, prometheusURL, nodeName string) (float64, error) {
	query := fmt.Sprintf(`100 * (1 - node_memory_MemAvailable_bytes{instance=~"%s.*"} / node_memory_MemTotal_bytes{instance=~"%s.*"})`, nodeName, nodeName)
	
	result, err := s.Query(ctx, prometheusURL, query)
	if err != nil {
		return 0, err
	}

	return s.extractFirstValue(result)
}

// CheckConnectivity checks if Prometheus is reachable
func (s *PrometheusService) CheckConnectivity(ctx context.Context, prometheusURL string) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.Query(ctx, prometheusURL, "up")
	return err == nil
}

// extractFirstValue extracts the first numeric value from a Prometheus result
func (s *PrometheusService) extractFirstValue(result *PrometheusResult) (float64, error) {
	if len(result.Data.Result) == 0 {
		return 0, nil
	}

	if len(result.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("unexpected result format")
	}

	valueStr, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("unexpected value type")
	}

	var value float64
	_, err := fmt.Sscanf(valueStr, "%f", &value)
	if err != nil {
		return 0, fmt.Errorf("failed to parse value: %w", err)
	}

	return value, nil
}

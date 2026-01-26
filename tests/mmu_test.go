package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/scross01/prometheus-klipper-exporter/collector"
)

// Test MMU data fetching and metric collection
func TestMMUMetrics(t *testing.T) {
	// Load test response
	responseData, err := os.ReadFile("mmu_response.json")
	if err != nil {
		t.Fatalf("Failed to read test response: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	}))
	defer server.Close()

	// Create collector instance using New function
	c := collector.New(context.Background(), server.URL[7:], []string{"mmu"}, "")

	// Test collectMMU function by checking generated metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		// Call Collect which should include MMU metrics since we enabled the mmu module
		c.Collect(ch)
		close(ch)
	}()

	// Collect all metrics
	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify we got some metrics (basic check)
	if len(metrics) == 0 {
		t.Error("No metrics were collected")
	}

	// Verify expected metrics are present by checking metric names
	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		metricNames[metric.Desc().String()] = true
	}

	// Check for key MMU metrics that should be present
	expectedMetrics := []string{
		"Desc{fqName: \"klipper_mmu_enabled\", help: \"MMU enabled state\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_mmu_num_gates\", help: \"Number of MMU gates\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_mmu_filament_loaded\", help: \"Filament loaded state\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_mmu_encoder_position_mm\", help: \"Encoder position in mm\", constLabels: {}, variableLabels: {}}",
	}

	for _, expectedMetric := range expectedMetrics {
		if !metricNames[expectedMetric] {
			t.Errorf("Expected metric %s not found", expectedMetric)
		}
	}
}

// Test MMU pre-gate sensors fetching
func TestMMUPreGateSensors(t *testing.T) {
	// Create a mock response for pre-gate sensors
	sensorResponse := map[string]interface{}{
		"result": map[string]interface{}{
			"status": map[string]interface{}{
				"filament_switch_sensor mmu_pre_gate_0": map[string]interface{}{
					"filament_detected": true,
					"enabled":           true,
				},
				"filament_switch_sensor mmu_pre_gate_1": map[string]interface{}{
					"filament_detected": false,
					"enabled":           true,
				},
			},
		},
	}

	sensorData, err := json.Marshal(sensorResponse)
	if err != nil {
		t.Fatalf("Failed to marshal sensor response: %v", err)
	}

	// Create test server for sensors
	sensorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(sensorData)
	}))
	defer sensorServer.Close()

	// Create collector instance
	c := collector.New(context.Background(), sensorServer.URL[7:], []string{"mmu"}, "")

	// Test by calling Collect which should include MMU sensor data collection
	ch := make(chan prometheus.Metric, 100)
	go func() {
		c.Collect(ch)
		close(ch)
	}()

	// Collect all metrics
	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Basic check that we got metrics
	if len(metrics) == 0 {
		t.Error("No metrics were collected from sensor test")
	}
}

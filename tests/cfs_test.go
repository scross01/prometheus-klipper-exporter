package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/scross01/prometheus-klipper-exporter/collector"
)

// Test CFS (Creality Filament System) data fetching and metric collection
func TestCFSMetrics(t *testing.T) {
	// Load test response
	responseData, err := os.ReadFile("cfs_response.json")
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
	c := collector.New(context.Background(), server.URL[7:], []string{"cfs"}, "")

	// Test collectCFS function by checking generated metrics
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

	if len(metrics) == 0 {
		t.Error("No metrics were collected")
	}

	// Index metric descriptions
	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		metricNames[metric.Desc().String()] = true
	}

	// Check for key CFS metrics that should be present (box, filament_rack)
	expectedMetrics := []string{
		"Desc{fqName: \"klipper_cfs_enabled\", help: \"CFS enabled state\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_cfs_active_slot\", help: \"Active slot index within the unit (A=0..D=3, -1 if none)\", constLabels: {}, variableLabels: {unit}}",
		"Desc{fqName: \"klipper_cfs_active_slot_info\", help: \"Active slot details (always 1)\", constLabels: {}, variableLabels: {unit,slot,material,color}}",
		"Desc{fqName: \"klipper_cfs_slot_info\", help: \"CFS slot details (always 1)\", constLabels: {}, variableLabels: {unit,slot,material,color,vendor}}",
		"Desc{fqName: \"klipper_cfs_unit_temperature_celsius\", help: \"CFS unit temperature in celsius\", constLabels: {}, variableLabels: {unit}}",
		"Desc{fqName: \"klipper_cfs_rack_loaded_info\", help: \"Filament currently loaded at the toolhead (always 1)\", constLabels: {}, variableLabels: {material,color}}",
	}

	for _, expectedMetric := range expectedMetrics {
		if !metricNames[expectedMetric] {
			t.Errorf("Expected metric %s not found", expectedMetric)
		}
	}

	// Disconnected units (T2/T3/T4 with state "None") must be skipped entirely, so no
	// emitted metric should carry a `unit` label value other than the connected "T1".
	for _, metric := range metrics {
		var m dto.Metric
		if err := metric.Write(&m); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}
		for _, label := range m.GetLabel() {
			if label.GetName() == "unit" && strings.HasPrefix(label.GetValue(), "T") && label.GetValue() != "T1" {
				t.Errorf("Unexpected metric for disconnected unit %s: %s", label.GetValue(), metric.Desc().String())
			}
		}
	}
}

package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/scross01/prometheus-klipper-exporter/collector"
)

func TestSystemInfoMetrics(t *testing.T) {
	fixture := `{
		"result": {
			"system_info": {
				"cpu_info": {
					"cpu_count": 4,
					"total_memory": 8192000,
					"memory_units": "kB"
				},
				"available_services": ["klipper", "moonraker", "webcamd"],
				"service_state": {
					"klipper": {
						"active": true,
						"active_state": "active",
						"sub_state": "running"
					},
					"moonraker": {
						"active": true,
						"active_state": "active",
						"sub_state": "running"
					},
					"webcamd": {
						"active": false,
						"active_state": "inactive",
						"sub_state": "dead"
					}
				}
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixture))
	}))
	defer server.Close()

	c := collector.New(context.Background(), server.URL[7:], []string{"system_info"}, "")

	ch := make(chan prometheus.Metric, 100)
	go func() {
		c.Collect(ch)
		close(ch)
	}()

	// Collect all metrics and their string representations
	var metricDescs []string
	for m := range ch {
		metricDescs = append(metricDescs, m.Desc().String())
	}

	// Verify we got metrics
	if len(metricDescs) == 0 {
		t.Fatal("No metrics were collected")
	}

	// Check for expected metric names
	metricNames := make(map[string]bool)
	for _, desc := range metricDescs {
		metricNames[desc] = true
	}

	expectedMetrics := []string{
		"Desc{fqName: \"klipper_system_cpu_count\", help: \"Klipper system CPU count.\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_service_available\", help: \"Klipper host service availability. Always 1 when present.\", constLabels: {}, variableLabels: {service}}",
		"Desc{fqName: \"klipper_service_state_info\", help: \"Klipper host service state.\", constLabels: {}, variableLabels: {service,state}}",
		"Desc{fqName: \"klipper_service_sub_state_info\", help: \"Klipper host service sub-state.\", constLabels: {}, variableLabels: {service,sub_state}}",
	}

	for _, expected := range expectedMetrics {
		if !metricNames[expected] {
			t.Errorf("Expected metric %s not found", expected)
		}
	}

	// Verify we have the right count of each metric
	cpuCountCount := 0
	availableCount := 0
	stateInfoCount := 0
	subStateInfoCount := 0
	for _, desc := range metricDescs {
		switch desc {
		case expectedMetrics[0]:
			cpuCountCount++
		case expectedMetrics[1]:
			availableCount++
		case expectedMetrics[2]:
			stateInfoCount++
		case expectedMetrics[3]:
			subStateInfoCount++
		}
	}

	if cpuCountCount != 1 {
		t.Errorf("Expected 1 klipper_system_cpu_count metric, got %d", cpuCountCount)
	}
	if availableCount != 3 {
		t.Errorf("Expected 3 klipper_service_available metrics, got %d", availableCount)
	}
	if stateInfoCount != 3 {
		t.Errorf("Expected 3 klipper_service_state_info metrics, got %d", stateInfoCount)
	}
	if subStateInfoCount != 3 {
		t.Errorf("Expected 3 klipper_service_sub_state_info metrics, got %d", subStateInfoCount)
	}
}

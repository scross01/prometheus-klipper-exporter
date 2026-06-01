package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/scross01/prometheus-klipper-exporter/collector"
)

// testSpoolmanHandler creates a handler that responds to /server/spoolman/status
// and /server/spoolman/proxy with the given fixtures.
func testSpoolmanHandler(statusFixture, proxyFixture string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/server/spoolman/status") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(statusFixture))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/server/spoolman/proxy") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(proxyFixture))
			return
		}
		http.NotFound(w, r)
	})
}

func TestSpoolmanMetricsV2(t *testing.T) {
	statusFixture := `{"result": {"spoolman_connected": true, "pending_reports": [], "spool_id": 3}}`

	proxyFixture := `{
		"result": {
			"response": [
				{
					"id": 1,
					"remaining_weight": 750.5,
					"used_weight": 249.5,
					"remaining_length": 250.2,
					"used_length": 83.1,
					"filament": {
						"name": "PLA+ Red",
						"material": "PLA",
						"color_hex": "#FF5733",
						"vendor": {"name": "HiPi.io"}
					}
				},
				{
					"id": 2,
					"remaining_weight": 500.0,
					"used_weight": 500.0,
					"remaining_length": 166.7,
					"used_length": 166.7,
					"filament": {
						"name": "PETG Black",
						"material": "PETG",
						"color_hex": "#000000",
						"vendor": {"name": "Fusion"}
					}
				}
			],
			"response_headers": {
				"Content-Type": "application/json",
				"X-Total-Count": "2"
			},
			"error": null
		}
	}`

	server := httptest.NewServer(testSpoolmanHandler(statusFixture, proxyFixture))
	defer server.Close()

	assertSpoolmanMetrics(t, server)
}

func TestSpoolmanMetricsV1(t *testing.T) {
	statusFixture := `{"result": {"spoolman_connected": true, "pending_reports": [], "spool_id": 3}}`

	proxyFixture := `{
		"result": [
			{
				"id": 1,
				"remaining_weight": 750.5,
				"used_weight": 249.5,
				"remaining_length": 250.2,
				"used_length": 83.1,
				"filament": {
					"name": "PLA+ Red",
					"material": "PLA",
					"color_hex": "#FF5733",
					"vendor": {"name": "HiPi.io"}
				}
			},
			{
				"id": 2,
				"remaining_weight": 500.0,
				"used_weight": 500.0,
				"remaining_length": 166.7,
				"used_length": 166.7,
				"filament": {
					"name": "PETG Black",
					"material": "PETG",
					"color_hex": "#000000",
					"vendor": {"name": "Fusion"}
				}
			}
		]
	}`

	server := httptest.NewServer(testSpoolmanHandler(statusFixture, proxyFixture))
	defer server.Close()

	assertSpoolmanMetrics(t, server)
}

func TestSpoolmanNoSpools(t *testing.T) {
	// Spoolman connected but no spools
	statusFixture := `{"result": {"spoolman_connected": true, "pending_reports": [], "spool_id": null}}`
	proxyFixture := `{"result": {"response": [], "response_headers": {"X-Total-Count": "0"}, "error": null}}`

	server := httptest.NewServer(testSpoolmanHandler(statusFixture, proxyFixture))
	defer server.Close()

	c := collector.New(context.Background(), server.URL[7:], []string{"spoolman"}, "")

	ch := make(chan prometheus.Metric, 100)
	go func() {
		c.Collect(ch)
		close(ch)
	}()

	var metricDescs []string
	for m := range ch {
		metricDescs = append(metricDescs, m.Desc().String())
	}

	// Should have status metrics (connected=1, active_spool_id=-1, pending_reports=0)
	expectedStatusMetrics := []string{
		"Desc{fqName: \"klipper_spoolman_connected\", help: \"Spoolman connection status (1=connected, 0=disconnected).\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_spoolman_active_spool_id\", help: \"Currently active spool ID (-1 if no spool is active).\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_spoolman_pending_reports\", help: \"Number of pending filament usage reports not yet sent to Spoolman.\", constLabels: {}, variableLabels: {}}",
	}

	for _, expected := range expectedStatusMetrics {
		found := false
		for _, desc := range metricDescs {
			if desc == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected status metric not found: %s", expected)
		}
	}

	// No per-spool metrics should be emitted for an empty spool list
	for _, desc := range metricDescs {
		if strings.Contains(desc, "klipper_spoolman_spool_info") ||
			strings.Contains(desc, "klipper_spoolman_remaining_weight") ||
			strings.Contains(desc, "klipper_spoolman_used_weight") ||
			strings.Contains(desc, "klipper_spoolman_remaining_length") ||
			strings.Contains(desc, "klipper_spoolman_used_length") {
			t.Errorf("Expected no per-spool metrics for empty spool list, got %s", desc)
		}
	}
}

func TestSpoolmanDisconnected(t *testing.T) {
	// Spoolman disconnected, has pending reports, no active spool
	statusFixture := `{"result": {"spoolman_connected": false, "pending_reports": [{"spool_id": 1, "filament_used": 10.5}], "spool_id": null}}`
	proxyFixture := `{"result": {"response": [], "response_headers": {"X-Total-Count": "0"}, "error": null}}`

	server := httptest.NewServer(testSpoolmanHandler(statusFixture, proxyFixture))
	defer server.Close()

	c := collector.New(context.Background(), server.URL[7:], []string{"spoolman"}, "")

	ch := make(chan prometheus.Metric, 100)
	go func() {
		c.Collect(ch)
		close(ch)
	}()

	var metricDescs []string
	for m := range ch {
		metricDescs = append(metricDescs, m.Desc().String())
	}

	metricNames := make(map[string]int)
	for _, desc := range metricDescs {
		metricNames[desc]++
	}

	// Status metrics should all be present
	expectedStatusMetrics := []string{
		"Desc{fqName: \"klipper_spoolman_connected\", help: \"Spoolman connection status (1=connected, 0=disconnected).\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_spoolman_active_spool_id\", help: \"Currently active spool ID (-1 if no spool is active).\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_spoolman_pending_reports\", help: \"Number of pending filament usage reports not yet sent to Spoolman.\", constLabels: {}, variableLabels: {}}",
	}
	for _, expected := range expectedStatusMetrics {
		if metricNames[expected] == 0 {
			t.Errorf("Expected status metric not found: %s", expected)
		}
	}

	// No per-spool metrics should be emitted
	for _, desc := range metricDescs {
		if strings.Contains(desc, "klipper_spoolman_spool_info") ||
			strings.Contains(desc, "klipper_spoolman_remaining_") ||
			strings.Contains(desc, "klipper_spoolman_used_") {
			t.Errorf("Expected no per-spool metrics when disconnected, got %s", desc)
		}
	}
}

func TestSpoolmanProxyError(t *testing.T) {
	// Error response from Spoolman proxy (v2 format)
	statusFixture := `{"result": {"spoolman_connected": true, "pending_reports": [], "spool_id": 3}}`
	proxyFixture := `{
		"result": {
			"response": null,
			"response_headers": null,
			"error": {
				"status_code": 503,
				"message": "Spoolman not connected."
			}
		}
	}`

	server := httptest.NewServer(testSpoolmanHandler(statusFixture, proxyFixture))
	defer server.Close()

	c := collector.New(context.Background(), server.URL[7:], []string{"spoolman"}, "")

	ch := make(chan prometheus.Metric, 100)
	go func() {
		c.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for m := range ch {
		metrics = append(metrics, m)
	}

	// Status metrics should still be emitted even if the proxy returns an error
	statusMetricsFound := map[string]bool{
		"klipper_spoolman_connected":      false,
		"klipper_spoolman_active_spool_id": false,
		"klipper_spoolman_pending_reports": false,
	}
	for _, m := range metrics {
		desc := m.Desc().String()
		for name := range statusMetricsFound {
			if strings.Contains(desc, name) {
				statusMetricsFound[name] = true
			}
		}
		// No per-spool metrics
		if strings.Contains(desc, "klipper_spoolman_spool_info") ||
			strings.Contains(desc, "klipper_spoolman_remaining_") ||
			strings.Contains(desc, "klipper_spoolman_used_") {
			t.Errorf("Expected no per-spool metrics on proxy error, got %s", desc)
		}
	}
	for name, found := range statusMetricsFound {
		if !found {
			t.Errorf("Expected status metric %s to be emitted", name)
		}
	}
}

// assertSpoolmanMetrics verifies that the server produces correct spoolman metrics
func assertSpoolmanMetrics(t *testing.T, server *httptest.Server) {
	t.Helper()

	c := collector.New(context.Background(), server.URL[7:], []string{"spoolman"}, "")

	ch := make(chan prometheus.Metric, 100)
	go func() {
		c.Collect(ch)
		close(ch)
	}()

	var metricDescs []string
	for m := range ch {
		metricDescs = append(metricDescs, m.Desc().String())
	}

	if len(metricDescs) == 0 {
		t.Fatal("No metrics were collected")
	}

	metricNames := make(map[string]bool)
	for _, desc := range metricDescs {
		metricNames[desc] = true
	}

	// Check status metrics are present
	expectedStatusMetrics := []string{
		"Desc{fqName: \"klipper_spoolman_connected\", help: \"Spoolman connection status (1=connected, 0=disconnected).\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_spoolman_active_spool_id\", help: \"Currently active spool ID (-1 if no spool is active).\", constLabels: {}, variableLabels: {}}",
		"Desc{fqName: \"klipper_spoolman_pending_reports\", help: \"Number of pending filament usage reports not yet sent to Spoolman.\", constLabels: {}, variableLabels: {}}",
	}
	for _, expected := range expectedStatusMetrics {
		if !metricNames[expected] {
			t.Errorf("Expected status metric %s not found", expected)
		}
	}

	// Check per-spool metrics are present
	expectedSpoolMetrics := []string{
		"Desc{fqName: \"klipper_spoolman_spool_info\", help: \"Spoolman spool information (always 1).\", constLabels: {}, variableLabels: {spool_id,filament_name,material,color,vendor}}",
		"Desc{fqName: \"klipper_spoolman_remaining_weight\", help: \"Remaining filament weight on the spool in grams.\", constLabels: {}, variableLabels: {spool_id}}",
		"Desc{fqName: \"klipper_spoolman_used_weight\", help: \"Used filament weight from the spool in grams.\", constLabels: {}, variableLabels: {spool_id}}",
		"Desc{fqName: \"klipper_spoolman_remaining_length\", help: \"Remaining filament length on the spool in millimetres.\", constLabels: {}, variableLabels: {spool_id}}",
		"Desc{fqName: \"klipper_spoolman_used_length\", help: \"Used filament length from the spool in millimetres.\", constLabels: {}, variableLabels: {spool_id}}",
	}
	for _, expected := range expectedSpoolMetrics {
		if !metricNames[expected] {
			t.Errorf("Expected spool metric %s not found", expected)
		}
	}

	// Verify correct counts (2 spools)
	spoolInfoCount := 0
	remainingWeightCount := 0
	usedWeightCount := 0
	remainingLengthCount := 0
	usedLengthCount := 0
	for _, desc := range metricDescs {
		switch desc {
		case expectedSpoolMetrics[0]:
			spoolInfoCount++
		case expectedSpoolMetrics[1]:
			remainingWeightCount++
		case expectedSpoolMetrics[2]:
			usedWeightCount++
		case expectedSpoolMetrics[3]:
			remainingLengthCount++
		case expectedSpoolMetrics[4]:
			usedLengthCount++
		}
	}

	if spoolInfoCount != 2 {
		t.Errorf("Expected 2 klipper_spoolman_spool_info metrics, got %d", spoolInfoCount)
	}
	if remainingWeightCount != 2 {
		t.Errorf("Expected 2 klipper_spoolman_remaining_weight metrics, got %d", remainingWeightCount)
	}
	if usedWeightCount != 2 {
		t.Errorf("Expected 2 klipper_spoolman_used_weight metrics, got %d", usedWeightCount)
	}
	if remainingLengthCount != 2 {
		t.Errorf("Expected 2 klipper_spoolman_remaining_length metrics, got %d", remainingLengthCount)
	}
	if usedLengthCount != 2 {
		t.Errorf("Expected 2 klipper_spoolman_used_length metrics, got %d", usedLengthCount)
	}
}

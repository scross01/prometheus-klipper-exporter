package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/scross01/prometheus-klipper-exporter/collector"
)

func TestSpoolmanMetrics(t *testing.T) {
	fixture := `[
		{
			"id": 1,
			"remaining_weight": 750.5,
			"used_weight": 249.5,
			"remaining_length": 250.2,
			"used_length": 83.1,
			"filament": {
				"name": "PLA+ Red",
				"material": "PLA",
				"color_hex": "#FF5733"
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
				"color_hex": "#000000"
			}
		}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixture))
	}))
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

	if len(metricDescs) == 0 {
		t.Fatal("No metrics were collected")
	}

	metricNames := make(map[string]bool)
	for _, desc := range metricDescs {
		metricNames[desc] = true
	}

	expectedMetrics := []string{
		"Desc{fqName: \"klipper_spoolman_spool_info\", help: \"Spoolman spool information (always 1).\", constLabels: {}, variableLabels: {spool_id,filament_name,material,color}}",
		"Desc{fqName: \"klipper_spoolman_remaining_weight\", help: \"Remaining filament weight on the spool in grams.\", constLabels: {}, variableLabels: {spool_id}}",
		"Desc{fqName: \"klipper_spoolman_used_weight\", help: \"Used filament weight from the spool in grams.\", constLabels: {}, variableLabels: {spool_id}}",
		"Desc{fqName: \"klipper_spoolman_remaining_length\", help: \"Remaining filament length on the spool in millimetres.\", constLabels: {}, variableLabels: {spool_id}}",
		"Desc{fqName: \"klipper_spoolman_used_length\", help: \"Used filament length from the spool in millimetres.\", constLabels: {}, variableLabels: {spool_id}}",
	}

	for _, expected := range expectedMetrics {
		if !metricNames[expected] {
			t.Errorf("Expected metric %s not found", expected)
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
		case expectedMetrics[0]:
			spoolInfoCount++
		case expectedMetrics[1]:
			remainingWeightCount++
		case expectedMetrics[2]:
			usedWeightCount++
		case expectedMetrics[3]:
			remainingLengthCount++
		case expectedMetrics[4]:
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

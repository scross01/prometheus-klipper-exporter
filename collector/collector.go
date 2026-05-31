package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type Collector struct {
	ctx     context.Context
	target  string
	modules []string
	apiKey  string
}

func New(ctx context.Context, target string, modules []string, apiKey string) *Collector {
	return &Collector{ctx: ctx, target: target, modules: modules, apiKey: apiKey}
}

// Describe implements Prometheus.Collector.
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Regex to match all invalid characters
var prometheusMetricNameInvalidCharactersRegex = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

// GetValidLabelName sanitizes a string to be used as a Prometheus label name
// It converts hyphens to underscores and removes all other invalid characters
func GetValidLabelName(str string) string {
	// convert hyphens to underscores and strip out all other invalid characters
	return prometheusMetricNameInvalidCharactersRegex.ReplaceAllString(strings.Replace(str, "-", "_", -1), "")
}

// A boolean cannot be directly converted to a number
func boolToFloat64(boolean bool) (value float64) {
	if boolean {
		value = 1
	}
	return value
}

// emitGauge is a convenience helper for emitting an unlabeled Gauge metric.
func (c Collector) emitGauge(ch chan<- prometheus.Metric, name, desc string, value float64) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(name, desc, nil, nil),
		prometheus.GaugeValue,
		value)
}

// emitCounter is a convenience helper for emitting an unlabeled Counter metric.
func (c Collector) emitCounter(ch chan<- prometheus.Metric, name, desc string, value float64) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(name, desc, nil, nil),
		prometheus.CounterValue,
		value)
}

// emitStateInfoMetric conditionally emits an info-style metric (Gauge=1) with a
// single label carrying the state value, only when the state is non-empty.
func emitStateInfoMetric(ch chan<- prometheus.Metric, metricName, description, labelName, stateValue string) {
	if stateValue != "" {
		desc := prometheus.NewDesc(metricName, description, []string{labelName}, nil)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1.0, stateValue)
	}
}

// fetchFromMoonraker performs an HTTP GET to the Moonraker API, JSON-unmarshals the
// response, and checks for a 200 status code.
func (c Collector) fetchFromMoonraker(urlPath string, response interface{}) error {
	url := "http://" + c.target + urlPath
	log.Debug("Collecting metrics from " + url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("unable to create HTTP request for %s: %w", url, err)
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-KEY", c.apiKey)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to complete HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d %s", res.StatusCode, res.Status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}

	if err := json.Unmarshal(data, response); err != nil {
		return fmt.Errorf("unable to unmarshal response to %T: %w", response, err)
	}

	return nil
}

// Collect implements Prometheus.Collector.
func (c Collector) Collect(ch chan<- prometheus.Metric) {

	// Process Stats (and Network Stats)
	if slices.Contains(c.modules, "process_stats") || slices.Contains(c.modules, "network_stats") {
		c.collectProcessAndNetworkStats(ch)
	}

	// Directory Information
	if slices.Contains(c.modules, "directory_info") {
		c.collectDirectoryInfo(ch)
	}

	// Job Queue
	if slices.Contains(c.modules, "job_queue") {
		c.collectJobQueue(ch)
	}

	// Job History
	if slices.Contains(c.modules, "history") {
		c.collectHistory(ch)
	}

	// Current Print from Job History
	if slices.Contains(c.modules, "history") {
		c.collectActivePrint(ch)
	}

	// Server Info
	if slices.Contains(c.modules, "server_info") {
		c.collectServerInfo(ch)
	}

	// System Info
	if slices.Contains(c.modules, "system_info") {
		c.collectSystemInfo(ch)
	}

	// Temperature Store
	// (deprecated since v0.8.0, use `printer_objects` instead)
	// (removed with warning in v0.14.0)
	if slices.Contains(c.modules, "temperature") {
		log.Errorf("Collecting `temperature` metrics for %s is no longer supported, use `printer_objects` instea", c.target)
	}

	// Printer Objects
	if slices.Contains(c.modules, "printer_objects") {
		c.collectPrinterObjects(ch)
	}

	// MMU (Multi-Material Unit) - Happy Hare - only if present
	if slices.Contains(c.modules, "mmu") {
		c.collectMMU(ch)
	}
}

// only return metric if current job status is in progress
func (c Collector) checkConditionStatusPrint(result MoonrakerHistoryCurrentPrintResponse, value float64) float64 {
	var valueToReturn float64 = 0
	if len(result.Result.Jobs) >= 1 && result.Result.Jobs[0].Status == "in_progress" {
		valueToReturn = value
	}
	return valueToReturn
}

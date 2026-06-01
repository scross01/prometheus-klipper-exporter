package collector

// Spoolman collector for Moonraker's Spoolman proxy integration
// https://moonraker.readthedocs.io/en/latest/external_api/integrations/#spoolman
//
// Moonraker proxies requests to the Spoolman API via POST /server/spoolman/proxy.
// The response is always wrapped in Moonraker's standard {"result": ...} envelope,
// where the inner result contains the Spoolman proxy response:
//
//   v2: {"result": {"response": [...], "error": null}}
//   v1: {"result": [...]}  (rare, older Moonraker)
//
// This module unwraps the result envelope and handles both formats.

import (
	"encoding/json"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// MoonrakerSpoolmanStatusResponse is the response from GET /server/spoolman/status
type MoonrakerSpoolmanStatusResponse struct {
	Result struct {
		SpoolmanConnected bool `json:"spoolman_connected"`
		PendingReports    []struct {
			SpoolID      int     `json:"spool_id"`
			FilamentUsed float64 `json:"filament_used"`
		} `json:"pending_reports"`
		SpoolID *int `json:"spool_id"`
	} `json:"result"`
}

// SpoolmanSpool represents a single spool returned by the Spoolman API
type SpoolmanSpool struct {
	ID              int     `json:"id"`
	RemainingWeight float64 `json:"remaining_weight"`
	UsedWeight      float64 `json:"used_weight"`
	RemainingLength float64 `json:"remaining_length"`
	UsedLength      float64 `json:"used_length"`
	Filament        struct {
		Name     string `json:"name"`
		Material string `json:"material"`
		ColorHex string `json:"color_hex"`
		Vendor   struct {
			Name string `json:"name"`
		} `json:"vendor"`
	} `json:"filament"`
}

// SpoolmanProxyRequest is the request body for POST /server/spoolman/proxy
type SpoolmanProxyRequest struct {
	RequestMethod string      `json:"request_method"`
	Path          string      `json:"path"`
	Query         *string     `json:"query"`
	Body          interface{} `json:"body"`
	UseV2Response bool        `json:"use_v2_response"`
}

// moonrakerProxyEnvelope is the standard Moonraker {"result": ...} wrapper
type moonrakerProxyEnvelope struct {
	Result json.RawMessage `json:"result"`
}

// spoolmanProxyV2Result is the v2 proxy result inside the result envelope
type spoolmanProxyV2Result struct {
	Response        json.RawMessage `json:"response"`
	ResponseHeaders *struct {
		ContentType string `json:"Content-Type"`
		XTotalCount string `json:"X-Total-Count"`
	} `json:"response_headers"`
	Error *struct {
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	} `json:"error"`
}

func (c Collector) collectSpoolman(ch chan<- prometheus.Metric) {
	// Collect Spoolman connection status and active spool info
	c.collectSpoolmanStatus(ch)

	// Collect per-spool metrics via proxy, excluding archived spools
	query := "archived=false"
	proxyBody := SpoolmanProxyRequest{
		RequestMethod: "GET",
		Path:          "/v1/spool",
		Query:         &query,
		Body:          nil,
		UseV2Response: true,
	}

	// Fetch raw response bytes so we can inspect the format
	var rawResponse json.RawMessage
	if err := c.fetchFromMoonrakerPost("/server/spoolman/proxy", proxyBody, &rawResponse); err != nil {
		log.Error(err)
		return
	}

	// Unwrap the standard Moonraker {"result": ...} envelope
	var envelope moonrakerProxyEnvelope
	if err := json.Unmarshal(rawResponse, &envelope); err != nil {
		log.Errorf("Unable to parse Moonraker response envelope for %s: %v", c.target, err)
		return
	}

	var spools []SpoolmanSpool

	if envelope.Result == nil {
		log.Errorf("Empty result from Spoolman proxy for %s", c.target)
		return
	}

	// Peek at the result format
	trimmed := envelope.Result
	if len(trimmed) == 0 {
		log.Errorf("Empty result from Spoolman proxy for %s", c.target)
		return
	}

	switch trimmed[0] {
	case '[':
		// v1 format inside result: raw JSON array of spools
		// {"result": [...]}
		if err := json.Unmarshal(envelope.Result, &spools); err != nil {
			log.Errorf("Unable to parse Spoolman proxy response for %s: %v", c.target, err)
			return
		}

	case '{':
		// v2 format inside result: {"response": [...], "error": null}
		var v2Result spoolmanProxyV2Result
		if err := json.Unmarshal(envelope.Result, &v2Result); err != nil {
			log.Errorf("Unable to parse Spoolman proxy result for %s: %v", c.target, err)
			return
		}

		// Check for Spoolman-side errors
		if v2Result.Error != nil {
			log.Errorf("Spoolman proxy error for %s (status %d): %s", c.target, v2Result.Error.StatusCode, v2Result.Error.Message)
			return
		}

		// No response field or null — no active spools
		if v2Result.Response == nil || string(v2Result.Response) == "null" {
			log.Infof("No spools found for %s", c.target)
			emitSpoolMetrics(ch, spools) // emits nothing
			return
		}

		// Parse the spool data
		if err := json.Unmarshal(v2Result.Response, &spools); err != nil {
			log.Errorf("Unable to parse spool data from proxy response for %s: %v", c.target, err)
			return
		}

	default:
		snippet := string(envelope.Result)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		log.Errorf("Unexpected Spoolman proxy result format for %s: %s", c.target, snippet)
		return
	}

	if len(spools) == 0 {
		log.Infof("No spools found for %s", c.target)
	}

	emitSpoolMetrics(ch, spools)
}

// collectSpoolmanStatus fetches Spoolman connection status and active spool info
// from the Moonraker Spoolman status endpoint.
func (c Collector) collectSpoolmanStatus(ch chan<- prometheus.Metric) {
	var status MoonrakerSpoolmanStatusResponse
	if err := c.fetchFromMoonraker("/server/spoolman/status", &status); err != nil {
		log.Error(err)
		return
	}

	// klipper_spoolman_connected — 1 if Moonraker has an active Spoolman connection
	c.emitGauge(ch, "klipper_spoolman_connected", "Spoolman connection status (1=connected, 0=disconnected).",
		boolToFloat64(status.Result.SpoolmanConnected))

	// klipper_spoolman_active_spool_id — current active spool ID (-1 if none)
	activeID := -1.0
	if status.Result.SpoolID != nil {
		activeID = float64(*status.Result.SpoolID)
	}
	c.emitGauge(ch, "klipper_spoolman_active_spool_id", "Currently active spool ID (-1 if no spool is active).", activeID)

	// klipper_spoolman_pending_reports — number of unsent filament usage reports
	c.emitGauge(ch, "klipper_spoolman_pending_reports", "Number of pending filament usage reports not yet sent to Spoolman.",
		float64(len(status.Result.PendingReports)))
}

// emitSpoolMetrics emits all spool-related Prometheus metrics for the given spools.
func emitSpoolMetrics(ch chan<- prometheus.Metric, spools []SpoolmanSpool) {
	// klipper_spoolman_spool_info{spool_id, filament_name, material, color, vendor} = 1
	spoolInfoLabels := []string{"spool_id", "filament_name", "material", "color", "vendor"}
	spoolInfoDesc := prometheus.NewDesc(
		"klipper_spoolman_spool_info",
		"Spoolman spool information (always 1).",
		spoolInfoLabels, nil,
	)

	// klipper_spoolman_remaining_weight{spool_id}
	remainingWeightDesc := prometheus.NewDesc(
		"klipper_spoolman_remaining_weight",
		"Remaining filament weight on the spool in grams.",
		[]string{"spool_id"}, nil,
	)

	// klipper_spoolman_used_weight{spool_id}
	usedWeightDesc := prometheus.NewDesc(
		"klipper_spoolman_used_weight",
		"Used filament weight from the spool in grams.",
		[]string{"spool_id"}, nil,
	)

	// klipper_spoolman_remaining_length{spool_id}
	remainingLengthDesc := prometheus.NewDesc(
		"klipper_spoolman_remaining_length",
		"Remaining filament length on the spool in millimetres.",
		[]string{"spool_id"}, nil,
	)

	// klipper_spoolman_used_length{spool_id}
	usedLengthDesc := prometheus.NewDesc(
		"klipper_spoolman_used_length",
		"Used filament length from the spool in millimetres.",
		[]string{"spool_id"}, nil,
	)

	for _, spool := range spools {
		spoolID := strconv.Itoa(spool.ID)

		// Info metric
		// Default vendor to "unknown" when not set (e.g. filament has no vendor association)
		vendorName := spool.Filament.Vendor.Name
		if vendorName == "" {
			vendorName = "unknown"
		}

		ch <- prometheus.MustNewConstMetric(
			spoolInfoDesc, prometheus.GaugeValue, 1.0,
			spoolID,
			GetValidLabelName(spool.Filament.Name),
			GetValidLabelName(spool.Filament.Material),
			GetValidLabelName(spool.Filament.ColorHex),
			GetValidLabelName(vendorName),
		)

		// Weight and length metrics
		ch <- prometheus.MustNewConstMetric(remainingWeightDesc, prometheus.GaugeValue, spool.RemainingWeight, spoolID)
		ch <- prometheus.MustNewConstMetric(usedWeightDesc, prometheus.GaugeValue, spool.UsedWeight, spoolID)
		ch <- prometheus.MustNewConstMetric(remainingLengthDesc, prometheus.GaugeValue, spool.RemainingLength, spoolID)
		ch <- prometheus.MustNewConstMetric(usedLengthDesc, prometheus.GaugeValue, spool.UsedLength, spoolID)
	}
}

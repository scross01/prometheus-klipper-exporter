package collector

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type MoonrakerServerInfoResponse struct {
	Result MoonrakerServerInfo `json:"result"`
}

type MoonrakerServerInfo struct {
	KlippyConnected  bool     `json:"klippy_connected"`
	KlippyState      string   `json:"klippy_state"`
	Components       []string `json:"components"`
	FailedComponents []string `json:"failed_components"`
	MoonrakerVersion string   `json:"moonraker_version"`
	APIVersion       []int    `json:"api_version"`
}

func (c Collector) collectServerInfo(ch chan<- prometheus.Metric) {
	log.Infof("Collecting server_info for %s", c.target)

	var result MoonrakerServerInfoResponse
	if err := c.fetchFromMoonraker("/server/info", &result); err != nil {
		log.Error(err)
		return
	}

	c.emitGauge(ch, "klipper_klippy_connected", "Whether Klippy is connected.", boolToFloat64(result.Result.KlippyConnected))
	emitStateInfoMetric(ch, "klipper_klippy_state_info", "The current state of Klippy.", "state", result.Result.KlippyState)

	for _, component := range result.Result.Components {
		emitStateInfoMetric(ch, "klipper_component_info", "A registered Moonraker component.", "component", component)
	}
	for _, component := range result.Result.FailedComponents {
		emitStateInfoMetric(ch, "klipper_component_failed_info", "A Moonraker component that failed to load.", "failed_component", component)
	}

	if result.Result.MoonrakerVersion != "" {
		emitStateInfoMetric(ch, "klipper_moonraker_version_info", "Moonraker version.", "version", result.Result.MoonrakerVersion)
	}

	if len(result.Result.APIVersion) > 0 {
		versionStr := formatAPIVersion(result.Result.APIVersion)
		emitStateInfoMetric(ch, "klipper_api_version_info", "Moonraker API version.", "version", versionStr)
	}
}

func formatAPIVersion(parts []int) string {
	strParts := make([]string, len(parts))
	for i, p := range parts {
		strParts[i] = fmt.Sprintf("%d", p)
	}
	return strings.Join(strParts, ".")
}

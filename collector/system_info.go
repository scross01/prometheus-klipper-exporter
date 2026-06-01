package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#get-system-info

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type MoonrakerSystemInfoQueryResponse struct {
	Result struct {
		SystemInfo struct {
			CpuInfo struct {
				CpuCount    int    `json:"cpu_count"`
				TotalMemory int    `json:"total_memory"`
				MemoryUnits string `json:"memory_units"`
			} `json:"cpu_info"`
			AvailableServices []string `json:"available_services"`
			ServiceState      map[string]struct {
				ActiveState string `json:"active_state"`
				SubState    string `json:"sub_state"`
			} `json:"service_state"`
		} `json:"system_info"`
	} `json:"result"`
}

func (c Collector) collectSystemInfo(ch chan<- prometheus.Metric) {
	var result MoonrakerSystemInfoQueryResponse
	if err := c.fetchFromMoonraker("/machine/system_info", &result); err != nil {
		log.Error(err)
		return
	}

	// CPU count
	c.emitGauge(ch, "klipper_system_cpu_count",
		"Klipper system CPU count.",
		float64(result.Result.SystemInfo.CpuInfo.CpuCount))

	// Iterate available_services and look up each in service_state to emit consistent metrics
	for _, service := range result.Result.SystemInfo.AvailableServices {
		labelName := GetValidLabelName(service)

		// Emit availability metric
		emitStateInfoMetric(ch, "klipper_service_available",
			"Klipper host service availability. Always 1 when present.",
			"service", labelName)

		// Look up the service in service_state
		if serviceStatus, exists := result.Result.SystemInfo.ServiceState[service]; exists {
			emitStateInfoMetric2(ch, "klipper_service_state_info",
				"Klipper host service state.",
				"service", labelName, "state", serviceStatus.ActiveState)
			emitStateInfoMetric2(ch, "klipper_service_sub_state_info",
				"Klipper host service sub-state.",
				"service", labelName, "sub_state", serviceStatus.SubState)
		} else {
			// Service is available but has no state — emit unknown sentinels
			emitStateInfoMetric2(ch, "klipper_service_state_info",
				"Klipper host service state.",
				"service", labelName, "state", "unknown")
			emitStateInfoMetric2(ch, "klipper_service_sub_state_info",
				"Klipper host service sub-state.",
				"service", labelName, "sub_state", "unknown")
		}
	}
}

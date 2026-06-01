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
				Active      bool   `json:"active"`
				ActiveState string `json:"active_state"`
				SubState    string `json:"sub_state"`
			} `json:"service_state"`
		} `json:"system_info"`
	} `json:"result"`
}

func (c Collector) collectSystemInfo(ch chan<- prometheus.Metric) {
	log.Infof("Collecting system_info for %s", c.target)

	var result MoonrakerSystemInfoQueryResponse
	if err := c.fetchFromMoonraker("/machine/system_info", &result); err != nil {
		log.Error(err)
		return
	}

	// CPU count
	c.emitGauge(ch, "klipper_system_cpu_count",
		"Klipper system CPU count.",
		float64(result.Result.SystemInfo.CpuInfo.CpuCount))

	// Emit klipper_service_available for each available service
	for _, service := range result.Result.SystemInfo.AvailableServices {
		emitStateInfoMetric(ch, "klipper_service_available",
			"Klipper host service availability. Always 1 when present.",
			"service", GetValidLabelName(service))
	}

	// Emit klipper_service_state_info and klipper_service_sub_state_info for each service
	for serviceName, serviceStatus := range result.Result.SystemInfo.ServiceState {
		labelName := GetValidLabelName(serviceName)

		if serviceStatus.ActiveState != "" {
			desc := prometheus.NewDesc("klipper_service_state_info",
				"Klipper host service state.",
				[]string{"service", "state"}, nil,
			)
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1.0, labelName, serviceStatus.ActiveState)
		}

		if serviceStatus.SubState != "" {
			desc := prometheus.NewDesc("klipper_service_sub_state_info",
				"Klipper host service sub-state.",
				[]string{"service", "sub_state"}, nil,
			)
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1.0, labelName, serviceStatus.SubState)
		}
	}
}

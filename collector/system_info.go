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

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_system_cpu_count", "Klipper system CPU count.", nil, nil),
		prometheus.GaugeValue,
		float64(result.Result.SystemInfo.CpuInfo.CpuCount))
}

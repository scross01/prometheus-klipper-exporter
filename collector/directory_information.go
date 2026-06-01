package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#get-directory-information

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type MoonrakerDirecotryInfoQueryResponse struct {
	Result struct {
		DiskUsage struct {
			Total int64 `json:"total"`
			Used  int64 `json:"used"`
			Free  int64 `json:"free"`
		} `json:"disk_usage"`
	} `json:"result"`
}

// collectDirectoryInfo
func (c Collector) collectDirectoryInfo(ch chan<- prometheus.Metric) {
	var result MoonrakerDirecotryInfoQueryResponse
	if err := c.fetchFromMoonraker("/server/files/directory?path=gcodes&extended=false", &result); err != nil {
		log.Error(err)
		return
	}

	c.emitGauge(ch, "klipper_disk_usage_total", "Klipper total disk space.", float64(result.Result.DiskUsage.Total))
	c.emitGauge(ch, "klipper_disk_usage_used", "Klipper used disk space.", float64(result.Result.DiskUsage.Used))
	c.emitGauge(ch, "klipper_disk_usage_available", "Klipper available disk space.", float64(result.Result.DiskUsage.Free))
}

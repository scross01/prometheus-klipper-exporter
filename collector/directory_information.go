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
	log.Infof("Collecting directory_info for %s", c.target)

	var result MoonrakerDirecotryInfoQueryResponse
	if err := c.fetchFromMoonraker("/server/files/directory?path=gcodes&extended=false", &result); err != nil {
		log.Error(err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_disk_usage_total", "Klipper total disk space.", nil, nil),
		prometheus.GaugeValue,
		float64(result.Result.DiskUsage.Total))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_disk_usage_used", "Klipper used disk space.", nil, nil),
		prometheus.GaugeValue,
		float64(result.Result.DiskUsage.Used))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_disk_usage_available", "Klipper available disk space.", nil, nil),
		prometheus.GaugeValue,
		float64(result.Result.DiskUsage.Free))
}

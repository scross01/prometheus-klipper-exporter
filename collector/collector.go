package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type collector struct {
	ctx    context.Context
	target string
	logger log.Logger
}

func New(ctx context.Context, target string, logger log.Logger) *collector {
	return &collector{ctx: ctx, target: target, logger: logger}
}

// Describe implements Prometheus.Collector.
func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect implements Prometheus.Collector.
func (c collector) Collect(ch chan<- prometheus.Metric) {

	result, err := fetchMoonrakerProcessStats(c.target)
	if err != nil {
		c.logger.Debug(err)
		return
	}

	memUnits := result.Result.MoonrakerStats[len(result.Result.MoonrakerStats)-1].MemUnits
	if memUnits != "kB" {
		log.Errorf("Unexpected units %s for Moonraker memory usage", memUnits)
	} else {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_moonraker_memory_kb", "Moonraker memory usage in Kb.", nil, nil),
		prometheus.GaugeValue,
		float64(result.Result.MoonrakerStats[len(result.Result.MoonrakerStats)-1].Memory))
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_moonraker_cpu_usage", "Moonraker CPU usage.", nil, nil),
		prometheus.GaugeValue,
		result.Result.MoonrakerStats[len(result.Result.MoonrakerStats)-1].CpuUsage)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_moonraker_websocket_connections", "Moonraker Websocket connection count.", nil, nil),
		prometheus.GaugeValue,
		float64(result.Result.WebsocketConnections))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_system_cpu_temp", "Klipper system CPU temperature in celsius.", nil, nil),
		prometheus.GaugeValue,
		result.Result.CpuTemp)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_system_cpu", "Klipper system CPU usage.", nil, nil),
		prometheus.GaugeValue,
		result.Result.SystemCpuUsage.Cpu)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_system_memory_total", "Klipper system total memory.", nil, nil),
		prometheus.GaugeValue,
		float64(result.Result.SystemMemory.Total))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_system_memory_available", "Klipper system available memory.", nil, nil),
		prometheus.GaugeValue,
		float64(result.Result.SystemMemory.Available))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_system_memory_used", "Klipper system used memory.", nil, nil),
		prometheus.GaugeValue,
		float64(result.Result.SystemMemory.Used))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_system_uptime", "Klipper system uptime.", nil, nil),
		prometheus.GaugeValue,
		result.Result.SystemUptime)

	result2, err := fetchMoonrakerDirectoryInfo(c.target)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_disk_usage_total", "Klipper total disk space.", nil, nil),
		prometheus.GaugeValue,
		float64(result2.Result.DiskUsage.Total))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_disk_usage_used", "Klipper used disk space.", nil, nil),
		prometheus.GaugeValue,
		float64(result2.Result.DiskUsage.Used))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_disk_usage_available", "Klipper available disk space.", nil, nil),
		prometheus.GaugeValue,
		float64(result2.Result.DiskUsage.Free))
	
	result3, err := fetchMoonrakerJobQueue(c.target)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_job_queue_length", "Klipper job queue length.", nil, nil),
		prometheus.GaugeValue,
		float64(len(result3.Result.QueuedJobs)))

	result4, err := fetchMoonrakerSystemInfo(c.target)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_system_cpu_count", "Klipper system CPU count.", nil, nil),
		prometheus.GaugeValue,
		float64(result4.Result.SystemInfo.CpuInfo.CpuCount))	
}

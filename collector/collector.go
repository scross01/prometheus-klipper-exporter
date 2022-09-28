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

	result, err := c.fetchMoonrakerProcessStats(c.target)
	if err != nil {
		c.logger.Debug(err)
		return
	}

	// Process Stats

	memUnits := result.Result.MoonrakerStats[len(result.Result.MoonrakerStats)-1].MemUnits
	if memUnits != "kB" {
		c.logger.Errorf("Unexpected units %s for Moonraker memory usage", memUnits)
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

	for key, element := range result.Result.Network {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_network_" + key + "_rx_bytes", "Klipper network recieved bytes.", nil, nil),
			prometheus.GaugeValue,
			float64(element.RxBytes))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_network_" + key + "_tx_bytes", "Klipper network transmitted bytes.", nil, nil),
			prometheus.GaugeValue,
			float64(element.TxBytes))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_network_" + key + "_rx_packets", "Klipper network recieved packets.", nil, nil),
			prometheus.GaugeValue,
			float64(element.RxPackets))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_network_" + key + "_tx_packets", "Klipper network transmitted packets.", nil, nil),
			prometheus.GaugeValue,
			float64(element.TxPackets))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_network_" + key + "_rx_errs", "Klipper network recieved errored packets.", nil, nil),
			prometheus.GaugeValue,
			float64(element.RxErrs))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_network_" + key + "_tx_errs", "Klipper network transmitted errored packets.", nil, nil),
			prometheus.GaugeValue,
			float64(element.TxErrs))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_network_" + key + "_rx_drop", "Klipper network recieved dropped packets.", nil, nil),
			prometheus.GaugeValue,
			float64(element.RxDrop))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_network_" + key + "_tx_drop", "Klipper network transmitted dropped packtes.", nil, nil),
			prometheus.GaugeValue,
			float64(element.TxDrop))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_network_" + key + "_bandwidth", "Klipper network bandwidth.", nil, nil),
			prometheus.GaugeValue,
			element.Bandwidth)
	}

	// Directory Information
	
	result2, err := c.fetchMoonrakerDirectoryInfo(c.target)
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
	
	// Job Queue

	result3, err := c.fetchMoonrakerJobQueue(c.target)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_job_queue_length", "Klipper job queue length.", nil, nil),
		prometheus.GaugeValue,
		float64(len(result3.Result.QueuedJobs)))

	// System Info

	result4, err := c.fetchMoonrakerSystemInfo(c.target)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_system_cpu_count", "Klipper system CPU count.", nil, nil),
		prometheus.GaugeValue,
		float64(result4.Result.SystemInfo.CpuInfo.CpuCount))	
}

package collector

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type collector struct {
	ctx    context.Context
	target string
	logger log.Logger
}

type MoonrakerProcessStatsQueryResponse struct {
	Result struct {
		MoonrakerStats       []MoonrakerProcStats             `json:"moonraker_stats"`
		CpuTemp              float64                          `json:"cpu_temp"`
		Network              map[string]MoonrakerNetworkStats `json:"network"`
		SystemCpuUsage       MoonrakerSystemCpuUsage          `json:"system_cpu_usage"`
		SystemMemory         MoonrakerSystemMemory            `json:"system_memory"`
		SystemUptime         float64                          `json:"system_uptime"`
		WebsocketConnections int                              `json:"websocket_connectsions"`
	} `json:"result"`
}

type MoonrakerProcStats struct {
	Time     float64 `json:"time"`
	CpuUsage float64 `json:"cpu_usage"`
	Memory   int     `json:"memory"`
	MemUnits string  `json:"mem_units"`
}

type MoonrakerNetworkStats struct {
	RxBytes   int     `json:"rx_bytes"`
	TxBytes   int     `json:"tx_bytes"`
	RxPackets int     `json:"rx_packets"`
	TxPackets int     `json:"tx_packets"`
	RxErrs    int     `json:"rx_errs"`
	TxErrs    int     `json:"tx_errs"`
	RxDrop    int     `json:"rx_drop"`
	TxDrop    int     `json:"tx_drop"`
	Bandwidth float64 `json:"bandwidth"`
}

type MoonrakerSystemCpuUsage struct {
	Cpu  float64 `json:"cpu"`
	Cpu0 float64 `json:"cpu0"`
	Cpu1 float64 `json:"cpu1"`
	Cpu2 float64 `json:"cpu2"`
	Cpu3 float64 `json:"cpu3"`
}

type MoonrakerSystemMemory struct {
	Total     int `json:"total"`
	Available int `json:"available"`
	Used      int `json:"used"`
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
	}

func fetchMoonrakerProcessStats(klipperHost string) (*MoonrakerProcessStatsQueryResponse, error) {
	var procStatsUrl = "http://" + klipperHost + "/machine/proc_stats"
	log.Info("Collecting metrics from " + procStatsUrl)
	res, err := http.Get(procStatsUrl)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var response MoonrakerProcessStatsQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	log.Info("Collected metrics from " + procStatsUrl)

	return &response, nil
}

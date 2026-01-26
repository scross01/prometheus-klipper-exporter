package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#get-moonraker-process-stats

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

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
	RxBytes   int64   `json:"rx_bytes"`
	TxBytes   int64   `json:"tx_bytes"`
	RxPackets int64   `json:"rx_packets"`
	TxPackets int64   `json:"tx_packets"`
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

func (c Collector) fetchMoonrakerProcessStats(klipperHost string, apiKey string) (*MoonrakerProcessStatsQueryResponse, error) {
	var url = "http://" + klipperHost + "/machine/proc_stats"
	log.Debug("Collecting metrics from " + url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP request for %s. %s", url, err)
	}
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to complete HTTP client request. %s", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read data from HTTP response. %s", err)
	}

	var response MoonrakerProcessStatsQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response data to %s. %s", reflect.TypeOf(response), err)
	}

	return &response, nil
}

func (c Collector) collectProcessAndNetworkStats(ch chan<- prometheus.Metric) bool {
	log.Infof("Collecting process_stats for %s", c.target)

	result, err := c.fetchMoonrakerProcessStats(c.target, c.apiKey)
	if err != nil {
		log.Error(err)
		return true
	}

	// Process Stats
	if slices.Contains(c.modules, "process_stats") {
		moonrakerStatsCount := len(result.Result.MoonrakerStats)
		if moonrakerStatsCount == 0 {
			log.Warn("Empty moonraker_stats in Process Stats response, skipping Memory and CPU usage stats")
		} else {
			memUnits := result.Result.MoonrakerStats[moonrakerStatsCount-1].MemUnits
			if memUnits != "kB" {
				log.Errorf("Unexpected units %s for Moonraker memory usage", memUnits)
			} else {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("klipper_moonraker_memory_kb", "Moonraker memory usage in Kb.", nil, nil),
					prometheus.GaugeValue,
					float64(result.Result.MoonrakerStats[moonrakerStatsCount-1].Memory))
			}

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_moonraker_cpu_usage", "Moonraker CPU usage.", nil, nil),
				prometheus.GaugeValue,
				result.Result.MoonrakerStats[moonrakerStatsCount-1].CpuUsage)
		}

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
			prometheus.CounterValue,
			result.Result.SystemUptime)
	}

	// Network Stats
	if slices.Contains(c.modules, "network_stats") {
		networkLabels := []string{"interface"}
		rxBytes := prometheus.NewDesc("klipper_network_rx_bytes", "Klipper network received bytes.", networkLabels, nil)
		txBytes := prometheus.NewDesc("klipper_network_tx_bytes", "Klipper network transmitted bytes.", networkLabels, nil)
		rxPackets := prometheus.NewDesc("klipper_network_rx_packets", "Klipper network received packets.", networkLabels, nil)
		txPackets := prometheus.NewDesc("klipper_network_tx_packets", "Klipper network transmitted packets.", networkLabels, nil)
		rxErrs := prometheus.NewDesc("klipper_network_rx_errs", "Klipper network received errored packets.", networkLabels, nil)
		txErrs := prometheus.NewDesc("klipper_network_tx_errs", "Klipper network transmitted errored packets.", networkLabels, nil)
		rxDrop := prometheus.NewDesc("klipper_network_rx_drop", "Klipper network received dropped packets.", networkLabels, nil)
		txDrop := prometheus.NewDesc("klipper_network_tx_drop", "Klipper network transmitted dropped packets.", networkLabels, nil)
		bandwidth := prometheus.NewDesc("klipper_network_bandwidth", "Klipper network bandwidth.", networkLabels, nil)
		for key, element := range result.Result.Network {
			interfaceName := GetValidLabelName(key)
			ch <- prometheus.MustNewConstMetric(
				rxBytes,
				prometheus.CounterValue,
				float64(element.RxBytes),
				interfaceName)
			ch <- prometheus.MustNewConstMetric(
				txBytes,
				prometheus.CounterValue,
				float64(element.TxBytes),
				interfaceName)
			ch <- prometheus.MustNewConstMetric(
				rxPackets,
				prometheus.CounterValue,
				float64(element.RxPackets),
				interfaceName)
			ch <- prometheus.MustNewConstMetric(
				txPackets,
				prometheus.CounterValue,
				float64(element.TxPackets),
				interfaceName)
			ch <- prometheus.MustNewConstMetric(
				rxErrs,
				prometheus.CounterValue,
				float64(element.RxErrs),
				interfaceName)
			ch <- prometheus.MustNewConstMetric(
				txErrs,
				prometheus.CounterValue,
				float64(element.TxErrs),
				interfaceName)
			ch <- prometheus.MustNewConstMetric(
				rxDrop,
				prometheus.CounterValue,
				float64(element.RxDrop),
				interfaceName)
			ch <- prometheus.MustNewConstMetric(
				txDrop,
				prometheus.CounterValue,
				float64(element.TxDrop),
				interfaceName)
			ch <- prometheus.MustNewConstMetric(
				bandwidth,
				prometheus.GaugeValue,
				element.Bandwidth,
				interfaceName)
		}
	}
	return false
}

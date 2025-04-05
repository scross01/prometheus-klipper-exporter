package collector

import (
	"context"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type Collector struct {
	ctx     context.Context
	target  string
	modules []string
	apiKey  string
}

func New(ctx context.Context, target string, modules []string, apiKey string) *Collector {
	return &Collector{ctx: ctx, target: target, modules: modules, apiKey: apiKey}
}

// Describe implements Prometheus.Collector.
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Regex to match all invalid characters
var prometheusMetricNameInvalidCharactersRegex = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

func getValidLabelName(str string) string {
	// convert hyphens to underscores and strip out all other invalid characters
	return prometheusMetricNameInvalidCharactersRegex.ReplaceAllString(strings.Replace(str, "-", "_", -1), "")
}

// A boolean cannot be directly converted to a number
func boolToFloat64(boolean bool) (value float64) {
	if boolean {
		value = 1
	}
	return value
}

// Collect implements Prometheus.Collector.
func (c Collector) Collect(ch chan<- prometheus.Metric) {

	// Process Stats (and Network Stats)
	if slices.Contains(c.modules, "process_stats") || slices.Contains(c.modules, "network_stats") {

		log.Infof("Collecting process_stats for %s", c.target)

		result, err := c.fetchMoonrakerProcessStats(c.target, c.apiKey)
		if err != nil {
			log.Error(err)
			return
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
				interfaceName := getValidLabelName(key)
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
	}

	// Directory Information
	if slices.Contains(c.modules, "directory_info") {
		log.Infof("Collecting directory_info for %s", c.target)
		result, err := c.fetchMoonrakerDirectoryInfo(c.target, c.apiKey)
		if err != nil {
			log.Error(err)
		} else {
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
	}

	// Job Queue
	if slices.Contains(c.modules, "job_queue") {
		log.Infof("Collecting job_queue for %s", c.target)
		result, err := c.fetchMoonrakerJobQueue(c.target, c.apiKey)
		if err != nil {
			log.Error(err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_job_queue_length", "Klipper job queue length.", nil, nil),
				prometheus.GaugeValue,
				float64(len(result.Result.QueuedJobs)))
		}
	}

	// Job History
	if slices.Contains(c.modules, "history") {
		log.Infof("Collecting history for %s", c.target)
		result, err := c.fetchMoonrakerHistory(c.target, c.apiKey)
		if err != nil {
			log.Error(err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_total_jobs", "Klipper number of total jobs.", nil, nil),
				prometheus.GaugeValue,
				float64(result.Result.JobTotals.Jobs))
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_total_time", "Klipper total time.", nil, nil),
				prometheus.GaugeValue,
				result.Result.JobTotals.TotalTime)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_total_print_time", "Klipper total print time.", nil, nil),
				prometheus.GaugeValue,
				result.Result.JobTotals.PrintTime)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_total_filament_used", "Klipper total meters of filament used.", nil, nil),
				prometheus.GaugeValue,
				result.Result.JobTotals.FilamentUsed)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_longest_job", "Klipper total longest job.", nil, nil),
				prometheus.GaugeValue,
				result.Result.JobTotals.LongestJob)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_longest_print", "Klipper total longest print.", nil, nil),
				prometheus.GaugeValue,
				result.Result.JobTotals.LongestPrint)
		}
	}

	// Current Print from Job History
	if slices.Contains(c.modules, "history") {
		log.Infof("Collecting active print for %s", c.target)
		result, err := c.fetchMoonrakerHistoryCurrent(c.target, c.apiKey)
		if err != nil {
			log.Error(err)
		} else {
			if len(result.Result.Jobs) < 1 {
				log.Info("No active print in Current Print repsonse, skipping current print metrics")
			} else {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("klipper_current_print_object_height", "Klipper current print object height", nil, nil),
					prometheus.GaugeValue,
					c.checkConditionStatusPrint(result, result.Result.Jobs[0].Metadata.ObjectHeight))
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("klipper_current_print_first_layer_height", "Klipper current print first layer height", nil, nil),
					prometheus.GaugeValue,
					c.checkConditionStatusPrint(result, result.Result.Jobs[0].Metadata.FirstLayerHeight))
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("klipper_current_print_layer_height", "Klipper current print layer height", nil, nil),
					prometheus.GaugeValue,
					c.checkConditionStatusPrint(result, result.Result.Jobs[0].Metadata.LayerHeight))
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("klipper_current_print_total_duration", "Klipper current print total duration", nil, nil),
					prometheus.GaugeValue,
					c.checkConditionStatusPrint(result, result.Result.Jobs[0].TotalDuration))
			}
		}
	}

	// System Info
	if slices.Contains(c.modules, "system_info") {
		log.Infof("Collecting system_info for %s", c.target)
		result, err := c.fetchMoonrakerSystemInfo(c.target, c.apiKey)
		if err != nil {
			log.Error(err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_system_cpu_count", "Klipper system CPU count.", nil, nil),
				prometheus.GaugeValue,
				float64(result.Result.SystemInfo.CpuInfo.CpuCount))
		}
	}

	// Temperature Store
	// (deprecated since v0.8.0, use `printer_objects` instead)
	if slices.Contains(c.modules, "temperature") {
		log.Infof("Collecting system_info for %s", c.target)
		result, err := c.fetchTemperatureData(c.target, c.apiKey)
		if err != nil {
			log.Error(err)
		} else {
			for k, v := range result.Result {
				item := strings.ReplaceAll(k, " ", "_")
				attributes := v.(map[string]interface{})
				for k1, v1 := range attributes {
					values := v1.([]interface{})
					label := strings.ReplaceAll(k1[0:len(k1)-1], " ", "_")
					ch <- prometheus.MustNewConstMetric(
						prometheus.NewDesc("klipper_"+item+"_"+label, "Klipper "+k+" "+label, nil, nil),
						prometheus.GaugeValue,
						values[len(values)-1].(float64))
				}
			}
		}
	}

	// Printer Objects
	if slices.Contains(c.modules, "printer_objects") {
		log.Infof("Collecting printer_objects for %s", c.target)
		result, err := c.fetchMoonrakerPrinterObjects(c.target, c.apiKey)
		if err != nil {
			log.Error(err)
		} else {

			// gcode_move
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_gcode_speed_factor", "Klipper gcode speed factor.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.GcodeMove.SpeedFactor)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_gcode_speed", "Klipper gcode speed.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.GcodeMove.Speed)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_gcode_extrude_factor", "Klipper gcode extrude factor.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.GcodeMove.ExtrudeFactor)

			// gcode position
			if len(result.Result.Status.GcodeMove.GcodePosition) < 4 {
				log.Warn("Unexpected number of Gcode Position values, skipping gcode position metrics")
			} else {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("klipper_gcode_position_x", "Klipper gcode position X axis.", nil, nil),
					prometheus.GaugeValue,
					result.Result.Status.GcodeMove.GcodePosition[0])
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("klipper_gcode_position_y", "Klipper gcode position Y axis.", nil, nil),
					prometheus.GaugeValue,
					result.Result.Status.GcodeMove.GcodePosition[1])
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("klipper_gcode_position_z", "Klipper gcode position Z axis.", nil, nil),
					prometheus.GaugeValue,
					result.Result.Status.GcodeMove.GcodePosition[2])
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("klipper_gcode_position_e", "Klipper gcode position for extruder.", nil, nil),
					prometheus.GaugeValue,
					result.Result.Status.GcodeMove.GcodePosition[3])
			}

			// mcu
			mcuLabels := []string{"mcu"}
			mcuAwake := prometheus.NewDesc("klipper_mcu_awake", "Klipper mcu awake.", mcuLabels, nil)
			mcuTaskAvg := prometheus.NewDesc("klipper_mcu_task_avg", "Klipper mcu task average.", mcuLabels, nil)
			mcuTaskStddev := prometheus.NewDesc("klipper_mcu_task_stddev", "Klipper mcu task standard deviation.", mcuLabels, nil)
			mcuWriteBytes := prometheus.NewDesc("klipper_mcu_write_bytes", "Klipper mcu write bytes.", mcuLabels, nil)
			mcuReadBytes := prometheus.NewDesc("klipper_mcu_read_bytes", "Klipper mcu read bytes.", mcuLabels, nil)
			mcuRetransmitBytes := prometheus.NewDesc("klipper_mcu_retransmit_bytes", "Klipper mcu retransmit bytes.", mcuLabels, nil)
			mcuInvalidBytes := prometheus.NewDesc("klipper_mcu_invalid_bytes", "Klipper mcu invalid bytes.", mcuLabels, nil)
			mcuSendSeq := prometheus.NewDesc("klipper_mcu_send_seq", "Klipper mcu send sequence.", mcuLabels, nil)
			mcuReceiveSeq := prometheus.NewDesc("klipper_mcu_receive_seq", "Klipper mcu receive sequence.", mcuLabels, nil)
			mcuRetransmitSeq := prometheus.NewDesc("klipper_mcu_retransmit_seq", "Klipper mcu retransmit sequence.", mcuLabels, nil)
			mcuSrtt := prometheus.NewDesc("klipper_mcu_srtt", "Klipper mcu smoothed round trip time.", mcuLabels, nil)
			mcuRttvar := prometheus.NewDesc("klipper_mcu_rttvar", "Klipper mcu round trip time variance.", mcuLabels, nil)
			mcuRto := prometheus.NewDesc("klipper_mcu_rto", "Klipper mcu retransmission timeouts.", mcuLabels, nil)
			mcuReadyBytes := prometheus.NewDesc("klipper_mcu_ready_bytes", "Klipper mcu ready bytes.", mcuLabels, nil)
			mcuStalledBytes := prometheus.NewDesc("klipper_mcu_stalled_bytes", "Klipper mcu stalled bytes.", mcuLabels, nil)
			mcuClockFrequency := prometheus.NewDesc("klipper_mcu_clock_frequency", "Klipper mcu clock frequency.", mcuLabels, nil)
			for mk, mv := range result.Result.Status.Mcus {
				sensorName := getValidLabelName(mk)
				ch <- prometheus.MustNewConstMetric(
					mcuAwake,
					prometheus.GaugeValue,
					mv.LastStats.McuAwake,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuTaskAvg,
					prometheus.GaugeValue,
					mv.LastStats.McuTaskAvg,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuTaskStddev,
					prometheus.GaugeValue,
					mv.LastStats.McuTaskStddev,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuWriteBytes,
					prometheus.GaugeValue,
					mv.LastStats.BytesWrite,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuReadBytes,
					prometheus.GaugeValue,
					mv.LastStats.BytesRead,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuRetransmitBytes,
					prometheus.GaugeValue,
					mv.LastStats.BytesRetransmit,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuInvalidBytes,
					prometheus.GaugeValue,
					mv.LastStats.BytesInvalid,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuSendSeq,
					prometheus.GaugeValue,
					mv.LastStats.SendSeq,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuReceiveSeq,
					prometheus.GaugeValue,
					mv.LastStats.ReceiveSeq,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuRetransmitSeq,
					prometheus.GaugeValue,
					mv.LastStats.RetransmitSeq,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuSrtt,
					prometheus.GaugeValue,
					mv.LastStats.Srtt,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuRttvar,
					prometheus.GaugeValue,
					mv.LastStats.Rttvar,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuRto,
					prometheus.GaugeValue,
					mv.LastStats.Rto,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuReadyBytes,
					prometheus.GaugeValue,
					mv.LastStats.ReadyBytes,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuStalledBytes,
					prometheus.GaugeValue,
					mv.LastStats.StalledBytes,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					mcuClockFrequency,
					prometheus.GaugeValue,
					mv.LastStats.Freq,
					sensorName)
			}

			// toolhead
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_toolhead_print_time", "Klipper toolhead print time.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Toolhead.PrintTime)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_toolhead_estimated_print_time", "Klipper estimated print time.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Toolhead.EstimatedPrintTime)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_toolhead_max_velocity", "Klipper toolhead max velocity.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Toolhead.MaxVelocity)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_toolhead_max_accel", "Klipper toolhead max acceleration.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Toolhead.MaxAccel)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_toolhead_max_accel_to_decel", "Klipper toolhead max acceleration to deceleration.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Toolhead.MaxAccelToDecel)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_toolhead_square_corner_velocity", "Klipper toolhead square corner velocity.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Toolhead.SquareCornerVelocity)

			// extruder
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_extruder_temperature", "Klipper extruder temperature.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Extruder.Temperature)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_extruder_target", "Klipper extruder target.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Extruder.Target)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_extruder_power", "Klipper extruder power.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Extruder.Power)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_extruder_pressure_advance", "Klipper extruder pressure advance.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Extruder.PressureAdvance)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_extruder_smooth_time", "Klipper extruder smooth time.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Extruder.SmoothTime)

			// heater_bed
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_heater_bed_temperature", "Klipper heater bed temperature.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.HeaterBed.Temperature)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_heater_bed_target", "Klipper heater bed target.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.HeaterBed.Target)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_heater_bed_power", "Klipper heater bed power.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.HeaterBed.Power)

			// fan
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_fan_speed", "Klipper fan speed.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Fan.Speed)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_fan_rpm", "Klipper fan rpm.", nil, nil),
				prometheus.GaugeValue,
				result.Result.Status.Fan.Rpm)

			// idle_timeout
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_printing_time", "The amount of time the printer has been in the Printing state.", nil, nil),
				prometheus.CounterValue,
				result.Result.Status.IdleTimeout.PrintingTime)

			// virtual_sdcard
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_print_file_progress", "The print progress reported as a percentage of the file read.", nil, nil),
				prometheus.CounterValue,
				result.Result.Status.VirtualSdCard.Progress)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_print_file_position", "The current file position in bytes.", nil, nil),
				prometheus.CounterValue,
				result.Result.Status.VirtualSdCard.FilePosition)

			// print_stats
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_print_total_duration", "The total time (in seconds) elapsed since a print has started.", nil, nil),
				prometheus.CounterValue,
				result.Result.Status.PrintStats.TotalDuration)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_print_print_duration", "The total time spent printing (in seconds).", nil, nil),
				prometheus.CounterValue,
				result.Result.Status.PrintStats.PrintDuration)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_print_filament_used", "The amount of filament used during the current print (in mm)..", nil, nil),
				prometheus.CounterValue,
				result.Result.Status.PrintStats.FilamentUsed)

			// display_status
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("klipper_print_gcode_progress", "The percentage of print progress, as reported by M73.", nil, nil),
				prometheus.CounterValue,
				result.Result.Status.DisplayStatus.Progress)

			// temperature_sensor
			temperatureSensorLabels := []string{"sensor"}
			temperatureSensor := prometheus.NewDesc("klipper_temperature_sensor_temperature", "The temperature of the temperature sensor", temperatureSensorLabels, nil)
			temperatureSensorMinTemp := prometheus.NewDesc("klipper_temperature_sensor_measured_min_temp", "The measured minimum temperature of the temperature sensor", temperatureSensorLabels, nil)
			temperatureSensorMaxTemp := prometheus.NewDesc("klipper_temperature_sensor_measured_max_temp", "The measured maximum temperature of the temperature sensor", temperatureSensorLabels, nil)
			for sk, sv := range result.Result.Status.TemperatureSensors {
				sensorName := getValidLabelName(sk)
				ch <- prometheus.MustNewConstMetric(
					temperatureSensor,
					prometheus.GaugeValue,
					sv.Temperature,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					temperatureSensorMinTemp,
					prometheus.GaugeValue,
					sv.MeasuredMinTemp,
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					temperatureSensorMaxTemp,
					prometheus.GaugeValue,
					sv.MeasuredMaxTemp,
					sensorName)
			}

			// temperature_fan
			fanLabels := []string{"fan"}
			fanSpeed := prometheus.NewDesc("klipper_temperature_fan_speed", "The speed of the temperature fan", fanLabels, nil)
			fanTemperature := prometheus.NewDesc("klipper_temperature_fan_temperature", "The temperature of the temperature fan", fanLabels, nil)
			fanTarget := prometheus.NewDesc("klipper_temperature_fan_target", "The target temperature for the temperature fan", fanLabels, nil)
			for fk, fv := range result.Result.Status.TemperatureFans {
				fanName := getValidLabelName(fk)
				ch <- prometheus.MustNewConstMetric(
					fanSpeed,
					prometheus.GaugeValue,
					fv.Speed,
					fanName)
				ch <- prometheus.MustNewConstMetric(
					fanTemperature,
					prometheus.GaugeValue,
					fv.Temperature,
					fanName)
				ch <- prometheus.MustNewConstMetric(
					fanTarget,
					prometheus.GaugeValue,
					fv.Target,
					fanName)
			}

			// temperature_probe
			temperatureProbeLabels := []string{"sensor"}
			temperatureProbe := prometheus.NewDesc("klipper_temperature_probe_temperature", "The temperature of the temperature probe", temperatureProbeLabels, nil)
			temperatureProbeMinTemp := prometheus.NewDesc("klipper_temperature_probe_measured_min_temp", "The measured minimum temperature of the temperature probe", temperatureProbeLabels, nil)
			temperatureProbeMaxTemp := prometheus.NewDesc("klipper_temperature_probe_measured_max_temp", "The measured maximum temperature of the temperature probe", temperatureProbeLabels, nil)
			temperatureProbeEstimatedExpansion := prometheus.NewDesc("klipper_temperature_probe_estimated_expansion", "The estimated of the temperature probe", temperatureProbeLabels, nil)
			for sk, sv := range result.Result.Status.TemperatureProbes {
				probeName := getValidLabelName(sk)
				ch <- prometheus.MustNewConstMetric(
					temperatureProbe,
					prometheus.GaugeValue,
					sv.Temperature,
					probeName)
				ch <- prometheus.MustNewConstMetric(
					temperatureProbeMinTemp,
					prometheus.GaugeValue,
					sv.MeasuredMinTemp,
					probeName)
				ch <- prometheus.MustNewConstMetric(
					temperatureProbeMaxTemp,
					prometheus.GaugeValue,
					sv.MeasuredMaxTemp,
					probeName)
				ch <- prometheus.MustNewConstMetric(
					temperatureProbeEstimatedExpansion,
					prometheus.GaugeValue,
					sv.EstimatedExpansion,
					probeName)
			}

			// output_pin
			pinLabels := []string{"pin"}
			pinValue := prometheus.NewDesc("klipper_output_pin_value", "The value of the output pin", pinLabels, nil)
			for k, v := range result.Result.Status.OutputPins {
				pinName := getValidLabelName(k)
				ch <- prometheus.MustNewConstMetric(
					pinValue,
					prometheus.GaugeValue,
					v.Value,
					pinName)
			}

			// fan_generic
			genericFanLabels := []string{"fan"}
			genericFanSpeed := prometheus.NewDesc("klipper_generic_fan_speed", "The speed of the generic fan", genericFanLabels, nil)
			genericFanRpm := prometheus.NewDesc("klipper_generic_fan_rpm", "The RPM of the generic fan", genericFanLabels, nil)
			for fk, fv := range result.Result.Status.GenericFans {
				fanName := getValidLabelName(fk)
				ch <- prometheus.MustNewConstMetric(
					genericFanSpeed,
					prometheus.GaugeValue,
					fv.Speed,
					fanName)
				ch <- prometheus.MustNewConstMetric(
					genericFanRpm,
					prometheus.GaugeValue,
					fv.Rpm,
					fanName)
			}

			// controller_fan
			controllerFanLabels := []string{"fan"}
			controllerFanSpeed := prometheus.NewDesc("klipper_controller_fan_speed", "The speed of the controller fan", controllerFanLabels, nil)
			controllerFanRpm := prometheus.NewDesc("klipper_controller_fan_rpm", "The RPM of the controller fan", controllerFanLabels, nil)
			for fk, fv := range result.Result.Status.ControllerFans {
				fanName := getValidLabelName(fk)
				ch <- prometheus.MustNewConstMetric(
					controllerFanSpeed,
					prometheus.GaugeValue,
					fv.Speed,
					fanName)
				ch <- prometheus.MustNewConstMetric(
					controllerFanRpm,
					prometheus.GaugeValue,
					fv.Rpm,
					fanName)
			}

			// filament_*_sensor
			filamentSensorLabels := []string{"sensor"}
			filamentSensorDetected := prometheus.NewDesc("klipper_filament_sensor_detected", "Whether filament presence is detected by the sensor", filamentSensorLabels, nil)
			filamentSensorEnabled := prometheus.NewDesc("klipper_filament_sensor_enabled", "Whether the filament sensor is enabled or not", filamentSensorLabels, nil)
			for k, v := range result.Result.Status.FilamentSensors {
				sensorName := getValidLabelName(k)
				ch <- prometheus.MustNewConstMetric(
					filamentSensorDetected,
					prometheus.GaugeValue,
					boolToFloat64(v.Detected),
					sensorName)
				ch <- prometheus.MustNewConstMetric(
					filamentSensorEnabled,
					prometheus.GaugeValue,
					boolToFloat64(v.Enabled),
					sensorName)
			}
		}
	}
}

// only return metric if current job status is in progress
func (c Collector) checkConditionStatusPrint(result *MoonrakerHistoryCurrentPrintResponse, value float64) float64 {
	var valueToReturn float64 = 0
	if len(result.Result.Jobs) >= 1 && result.Result.Jobs[0].Status == "in_progress" {
		valueToReturn = value
	}
	return valueToReturn
}

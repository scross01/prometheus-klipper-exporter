package collector

import (
	"context"
	"regexp"
	"strconv"
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

func b2f(cond bool) float64 {
	if cond {
		return 1.0
	}
	return 0.0
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
		result, _ := c.fetchMoonrakerDirectoryInfo(c.target, c.apiKey)
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

	// Job Queue
	if slices.Contains(c.modules, "job_queue") {
		log.Infof("Collecting job_queue for %s", c.target)
		result, _ := c.fetchMoonrakerJobQueue(c.target, c.apiKey)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_job_queue_length", "Klipper job queue length.", nil, nil),
			prometheus.GaugeValue,
			float64(len(result.Result.QueuedJobs)))
	}

	// Job History
	if slices.Contains(c.modules, "history") {
		log.Infof("Collecting history for %s", c.target)
		result, _ := c.fetchMoonrakerHistory(c.target, c.apiKey)
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

	// Current Print from Job History
	if slices.Contains(c.modules, "history") {
		log.Infof("Collecting active print for %s", c.target)
		result, _ := c.fetchMoonrakerHistoryCurrent(c.target, c.apiKey)
		if len(result.Result.Jobs) >= 1 {
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

	// System Info
	if slices.Contains(c.modules, "system_info") {
		log.Infof("Collecting system_info for %s", c.target)
		result, _ := c.fetchMoonrakerSystemInfo(c.target, c.apiKey)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_system_cpu_count", "Klipper system CPU count.", nil, nil),
			prometheus.GaugeValue,
			float64(result.Result.SystemInfo.CpuInfo.CpuCount))
	}

	// Temperature Store
	// (deprecated since v0.8.0, use `printer_objects` instead)
	if slices.Contains(c.modules, "temperature") {
		log.Infof("Collecting system_info for %s", c.target)
		result, _ := c.fetchTemperatureData(c.target, c.apiKey)

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

	// Printer Objects
	if slices.Contains(c.modules, "printer_objects") {
		log.Infof("Collecting printer_objects for %s", c.target)
		result, _ := c.fetchMoonrakerPrinterObjects(c.target, c.apiKey)

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
		if len(result.Result.Status.GcodeMove.GcodePosition) >= 4 {
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
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_awake", "Klipper mcu awake.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.McuAwake)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_write_bytes", "Klipper mcu write bytes.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.BytesWrite)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_read_bytes", "Klipper mcu read bytes.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.BytesRead)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_retransmit_bytes", "Klipper mcu retransmit bytes.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.BytesRetransmit)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_invalid_bytes", "Klipper mcu invalid bytes.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.BytesInvalid)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_send_seq", "Klipper mcu send sequence.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.SendSeq)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_receive_seq", "Klipper mcu receive sequence.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.ReceiveSeq)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_retransmit_seq", "Klipper mcu retransmit sequence.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.RetransmitSeq)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_srtt", "Klipper mcu smoothed round trip time.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.Srtt)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_rttvar", "Klipper mcu round trip time variance.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.Rttvar)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_rto", "Klipper mcu retransmission timeouts.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.Rto)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_ready_bytes", "Klipper mcu ready bytes.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.ReadyBytes)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_stalled_bytes", "Klipper mcu stalled bytes.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.StalledBytes)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mcu_clock_frequency", "Klipper mcu clock frequency.", nil, nil),
			prometheus.GaugeValue,
			result.Result.Status.Mcu.LastStats.Freq)

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
	}

	if slices.Contains(c.modules, "spoolman") {
		log.Infof("Collecting spoolman status for %s", c.target)
		result, _ := c.fetchSpoolmanStatus(c.target, c.apiKey)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_spoolman_status", "Klipper spoolman connected status", nil, nil),
			prometheus.GaugeValue,
			b2f(result.Result.SpoolmanConnected))

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_spoolman_current_spool", "Klipper current spool", []string{"spool_id"}, nil),
			prometheus.GaugeValue,
			1.0,
			strconv.Itoa(result.Result.SpoolId))

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

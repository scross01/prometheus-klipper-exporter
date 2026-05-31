package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#query-printer-object-status

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
)

type PrinterObjectResponse struct {
	Result struct {
		Status PrinterObjectStatus `json:"status"`
	} `json:"result"`
}

type PrinterObjectStatus struct {
	Webhooks      PrinterObjectWebhooks      `json:"webhooks"`
	PauseResume   PrinterObjectPauseResume   `json:"pause_resume"`
	GcodeMove     PrinterObjectGcodeMove     `json:"gcode_move"`
	Toolhead      PrinterObjectToolhead      `json:"toolhead"`
	Extruder      PrinterObjectExtruder      `json:"extruder"`
	HeaterBed     PrinterObjectHeaterBed     `json:"heater_bed"`
	Fan           PrinterObjectFan           `json:"fan"`
	IdleTimeout   PrinterObjectIdleTimeout   `json:"idle_timeout"`
	VirtualSdCard PrinterObjectVirtualSdCard `json:"virtual_sdcard"`
	PrintStats    PrinterObjectPrintStats    `json:"print_stats"`
	DisplayStatus PrinterObjectDisplayStatus `json:"display_status"`
	// dynamic sensor attributes populated using custom unmarsaling
	Mcus               map[string]PrinterObjectMcu
	TemperatureSensors map[string]PrinterObjectTemperatureSensor
	TemperatureFans    map[string]PrinterObjectTemperatureFan
	TemperatureProbes  map[string]PrinterObjectTemperatureProbe
	OutputPins         map[string]PrinterObjectOutputPin
	GenericFans        map[string]PrinterObjectFan
	ControllerFans     map[string]PrinterObjectFan
	HeaterFans         map[string]PrinterObjectFan
	FilamentSensors    map[string]PrinterObjectFilamentSensor
	GenericHeaters     map[string]PrinterObjectHeater
	TmcSensors         map[string]PrinterObjectTmc
}

type PrinterObjectMcu struct {
	LastStats struct {
		McuAwake        float64 `mapstructure:"mcu_awake"`
		McuTaskAvg      float64 `mapstructure:"mcu_task_avg"`
		McuTaskStddev   float64 `mapstructure:"mcu_task_stddev"`
		BytesWrite      float64 `mapstructure:"bytes_write"`
		BytesRead       float64 `mapstructure:"bytes_read"`
		BytesRetransmit float64 `mapstructure:"bytes_retransmit"`
		BytesInvalid    float64 `mapstructure:"bytes_invalid"`
		SendSeq         float64 `mapstructure:"send_seq"`
		ReceiveSeq      float64 `mapstructure:"receive_seq"`
		RetransmitSeq   float64 `mapstructure:"retransmit_seq"`
		Srtt            float64 `mapstructure:"srtt"`
		Rttvar          float64 `mapstructure:"rttvar"`
		Rto             float64 `mapstructure:"rto"`
		ReadyBytes      float64 `mapstructure:"ready_bytes"`
		StalledBytes    float64 `mapstructure:"stalled_bytes"`
		Freq            float64 `mapstructure:"freq"`
	} `mapstructure:"last_stats"`
}

type PrinterObjectGcodeMove struct {
	SpeedFactor   float64   `json:"speed_factor"`
	Speed         float64   `json:"speed"`
	ExtrudeFactor float64   `json:"extrude_factor"`
	GcodePosition []float64 `json:"gcode_position"`
}

const gcodeMoveQuery string = "gcode_move=speed_factor,speed,extrude_factor,gcode_position"

type PrinterObjectToolhead struct {
	PrintTime            float64 `json:"print_time"`
	EstimatedPrintTime   float64 `json:"estimated_print_time"`
	MaxVelocity          float64 `json:"max_velocity"`
	MaxAccel             float64 `json:"max_accel"`
	MaxAccelToDecel      float64 `json:"max_accel_to_decel"`
	SquareCornerVelocity float64 `json:"square_corner_velocity"`
	HomedAxes            string  `json:"homed_axes"`
	Stalls               float64 `json:"stalls"`
}

const toolheadQuery string = "toolhead=print_time,estimated_print_time,max_velocity,max_accel,max_accel_to_decel,square_corner_velocity,homed_axes,stalls"

type PrinterObjectExtruder struct {
	Temperature     float64 `json:"temperature"`
	Target          float64 `json:"target"`
	Power           float64 `json:"power"`
	PressureAdvance float64 `json:"pressure_advance"`
	SmoothTime      float64 `json:"smooth_time"`
}

const extruderQuery string = "extruder"

type PrinterObjectHeaterBed struct {
	Temperature float64 `json:"temperature"`
	Target      float64 `json:"target"`
	Power       float64 `json:"power"`
}

const heaterBedQuery = "heater_bed"

type PrinterObjectHeater struct {
	Temperature float64 `json:"temperature"`
	Target      float64 `json:"target"`
	Power       float64 `json:"power"`
}

const heaterQuery = "heaters"

type PrinterObjectFan struct {
	Speed float64  `json:"speed"`
	Rpm   *float64 `json:"rpm"`
}

const fanQuery = "fan"

type PrinterObjectWebhooks struct {
	State string `json:"state"`
}

const webhooksQuery = "webhooks"

type PrinterObjectPauseResume struct {
	IsPaused bool `json:"is_paused"`
}

const pauseResumeQuery = "pause_resume"

type PrinterObjectIdleTimeout struct {
	State        string  `json:"state"`
	PrintingTime float64 `json:"printing_time"`
}

const idleTimeoutQuery = "idle_timeout"

type PrinterObjectVirtualSdCard struct {
	Progress     float64 `json:"progress"`
	IsActive     bool    `json:"is_active"`
	FilePosition float64 `json:"file_position"`
}

const virtualSdCardQuery = "virtual_sdcard"

type PrinterObjectPrintStats struct {
	State         string  `json:"state"`
	TotalDuration float64 `json:"total_duration"`
	PrintDuration float64 `json:"print_duration"`
	FilamentUsed  float64 `json:"filament_used"`
}

const printStatsQuery = "print_stats=state,total_duration,print_duration,filament_used"

type PrinterObjectDisplayStatus struct {
	Progress float64 `json:"progress"`
}

const displayStatusQuery = "display_status"

type PrinterObjectTemperatureSensor struct {
	Temperature     float64 `mapstructure:"temperature"`
	MeasuredMinTemp float64 `mapstructure:"measured_min_temp"`
	MeasuredMaxTemp float64 `mapstructure:"measured_max_temp"`
}

type PrinterObjectTemperatureFan struct {
	Speed       float64  `mapstructure:"speed"`
	Temperature float64  `mapstructure:"temperature"`
	Target      float64  `mapstructure:"target"`
	Rpm         *float64 `mapstructure:"rpm"`
}

type PrinterObjectTemperatureProbe struct {
	Temperature        float64 `mapstructure:"temperature"`
	MeasuredMinTemp    float64 `mapstructure:"measured_min_temp"`
	MeasuredMaxTemp    float64 `mapstructure:"measured_max_temp"`
	EstimatedExpansion float64 `mapstructure:"estimated_expansion"`
}

type PrinterObjectOutputPin struct {
	Value float64 `mapstructure:"value"`
}

type PrinterObjectFilamentSensor struct {
	Detected bool `mapstructure:"filament_detected"`
	Enabled  bool `mapstructure:"enabled"`
}

type PrinterObjectTmcDrvStatus struct {
	// Empty struct because we only use it to check if tmc driver is enabled
}

type PrinterObjectTmc struct {
	RunCurrent  float64                    `mapstructure:"run_current"`
	Temperature *float64                   `mapstructure:"temperature"`
	DrvStatus   *PrinterObjectTmcDrvStatus `mapstructure:"drv_status"`
}

type _PrinterObjectStatus PrinterObjectStatus

func parseFanType(prefix string, key string, raw interface{}, target map[string]PrinterObjectFan) {
	if strings.HasPrefix(key, prefix) {
		name := strings.Replace(key, prefix+" ", "", 1)
		var value PrinterObjectFan
		mapstructure.Decode(raw, &value)
		target[name] = value
	}
}

func (f *PrinterObjectStatus) UnmarshalJSON(bs []byte) (err error) {
	status := _PrinterObjectStatus{}

	if err = json.Unmarshal(bs, &status); err == nil {
		*f = PrinterObjectStatus(status)
	}

	m := make(map[string]interface{})

	if err = json.Unmarshal(bs, &m); err == nil {
		// find `temperature_sensor` `temperature_fan` and `output_pin` items
		// and store in a map keyed by sensor name
		microcontrollers := make(map[string]PrinterObjectMcu)
		temperatureSensors := make(map[string]PrinterObjectTemperatureSensor)
		temperatureFans := make(map[string]PrinterObjectTemperatureFan)
		temperatureProbes := make(map[string]PrinterObjectTemperatureProbe)
		outputPins := make(map[string]PrinterObjectOutputPin)
		genericFans := make(map[string]PrinterObjectFan)
		controllerFans := make(map[string]PrinterObjectFan)
		heaterFans := make(map[string]PrinterObjectFan)
		filamentSensors := make(map[string]PrinterObjectFilamentSensor)
		genericHeaters := make(map[string]PrinterObjectHeater)
		tmcSensors := make(map[string]PrinterObjectTmc)
		for k, v := range m {
			// find mcus
			mcuMatch := mcuRegex.FindStringSubmatch(k)
			if mcuMatch != nil {
				key := k
				groupMatchIndex := mcuRegex.SubexpIndex("label")
				if mcuMatch[groupMatchIndex] != "" {
					key = strings.TrimSpace(mcuMatch[groupMatchIndex])
				}
				value := PrinterObjectMcu{}
				mapstructure.Decode(v, &value)
				microcontrollers[key] = value
			}
			if strings.HasPrefix(k, "temperature_sensor") {
				key := strings.Replace(k, "temperature_sensor ", "", 1)
				value := PrinterObjectTemperatureSensor{}
				mapstructure.Decode(v, &value)
				temperatureSensors[key] = value
			}
			if strings.HasPrefix(k, "temperature_fan") {
				key := strings.Replace(k, "temperature_fan ", "", 1)
				value := PrinterObjectTemperatureFan{}
				mapstructure.Decode(v, &value)
				temperatureFans[key] = value
			}
			if strings.HasPrefix(k, "temperature_probe") {
				key := strings.Replace(k, "temperature_probe ", "", 1)
				value := PrinterObjectTemperatureProbe{}
				mapstructure.Decode(v, &value)
				temperatureProbes[key] = value
			}
			if strings.HasPrefix(k, "output_pin") {
				key := strings.Replace(k, "output_pin ", "", 1)
				value := PrinterObjectOutputPin{}
				mapstructure.Decode(v, &value)
				outputPins[key] = value
			}
			parseFanType("fan_generic", k, v, genericFans)
			parseFanType("controller_fan", k, v, controllerFans)
			parseFanType("heater_fan", k, v, heaterFans)
			if filamentSensorRegex.MatchString(k) {
				key := strings.TrimSpace(filamentSensorRegex.ReplaceAllString(k, ""))
				value := PrinterObjectFilamentSensor{}
				mapstructure.Decode(v, &value)
				filamentSensors[key] = value
			}
			if strings.HasPrefix(k, "heater_generic") {
				key := strings.Replace(k, "heater_generic ", "", 1)
				value := PrinterObjectHeater{}
				mapstructure.Decode(v, &value)
				genericHeaters[key] = value
			}
			if strings.HasPrefix(k, "tmc") {
				value := PrinterObjectTmc{}
				mapstructure.Decode(v, &value)
				tmcSensors[k] = value
			}
		}
		f.Mcus = microcontrollers
		f.TemperatureSensors = temperatureSensors
		f.TemperatureFans = temperatureFans
		f.TemperatureProbes = temperatureProbes
		f.OutputPins = outputPins
		f.GenericFans = genericFans
		f.ControllerFans = controllerFans
		f.HeaterFans = heaterFans
		f.FilamentSensors = filamentSensors
		f.GenericHeaters = genericHeaters
		f.TmcSensors = tmcSensors
	}
	return err
}

type PrinterObjectsList struct {
	Result struct {
		Objects []string `json:"objects"`
	} `json:"result"`
}

var (
	filamentSensorRegex      *regexp.Regexp = regexp.MustCompile("^filament_(switch|motion)_sensor ")
	mcuRegex                 *regexp.Regexp = regexp.MustCompile("^mcu(?P<label> [a-zA-Z0-9_]+)?")
	customSensorsMu          sync.Mutex
	customMicrocontrollers   map[string][]string   = make(map[string][]string)
	customTemperatureSensors map[string][]string   = make(map[string][]string)
	customTemperatureFans    map[string][]string   = make(map[string][]string)
	customTemperatureProbes  map[string][]string   = make(map[string][]string)
	customOutputPins         map[string][]string   = make(map[string][]string)
	customGenericFans        map[string][]string   = make(map[string][]string)
	customControllerFans     map[string][]string   = make(map[string][]string)
	customHeaterFans         map[string][]string   = make(map[string][]string)
	customFilamentSensors    map[string][][]string = make(map[string][][]string)
	customGenericHeaters     map[string][]string   = make(map[string][]string)
	customTmcSensors         map[string][]string   = make(map[string][]string)
)

// fetchCustomSensors queries klipper for the complete list and printer objects and
// returns the subset of `temperature_sensor`, `temperature_fan`, `output_pin`,
// `fan_generic`, `controller_fan`, and `filament_*_sensor` objects that have custom names.
func (c Collector) fetchCustomSensors() (*[]string, *[]string, *[]string, *[]string, *[]string, *[]string, *[]string, *[]string, *[][]string, *[]string, *[]string, error) {
	var response PrinterObjectsList
	if err := c.fetchFromMoonraker("/printer/objects/list", &response); err != nil {
		log.Error(err)
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	microcontrollers := []string{}
	temperatureSensors := []string{}
	temperatureFans := []string{}
	temperatureProbes := []string{}
	outputPins := []string{}
	genericFans := []string{}
	controllerFans := []string{}
	heaterFans := []string{}
	filamentSensors := [][]string{}
	genericHeaters := []string{}
	tmcSensors := []string{}
	for o := range response.Result.Objects {
		// find mcus
		mcuMatch := mcuRegex.FindStringSubmatch(response.Result.Objects[o])
		if mcuMatch != nil {
			groupMatchIndex := mcuRegex.SubexpIndex("label")
			if mcuMatch[groupMatchIndex] == "" {
				microcontrollers = append(microcontrollers, response.Result.Objects[o])
			} else {
				microcontrollers = append(microcontrollers, strings.TrimSpace(mcuMatch[groupMatchIndex]))
			}
		}
		// find temperature_sensor
		if strings.HasPrefix(response.Result.Objects[o], "temperature_sensor ") {
			temperatureSensors = append(temperatureSensors, strings.Replace(response.Result.Objects[o], "temperature_sensor ", "", 1))
		}
		// find temperature_fan
		if strings.HasPrefix(response.Result.Objects[o], "temperature_fan ") {
			temperatureFans = append(temperatureFans, strings.Replace(response.Result.Objects[o], "temperature_fan ", "", 1))
		}
		// find temperature_probe
		if strings.HasPrefix(response.Result.Objects[o], "temperature_probe ") {
			temperatureProbes = append(temperatureProbes, strings.Replace(response.Result.Objects[o], "temperature_probe ", "", 1))
		}
		// find output_pin
		if strings.HasPrefix(response.Result.Objects[o], "output_pin ") {
			outputPins = append(outputPins, strings.Replace(response.Result.Objects[o], "output_pin ", "", 1))
		}
		// find fan_generic
		if strings.HasPrefix(response.Result.Objects[o], "fan_generic ") {
			genericFans = append(genericFans, strings.Replace(response.Result.Objects[o], "fan_generic ", "", 1))
		}
		// find controller_fan
		if strings.HasPrefix(response.Result.Objects[o], "controller_fan ") {
			controllerFans = append(controllerFans, strings.Replace(response.Result.Objects[o], "controller_fan ", "", 1))
		}
		// find heater_fan
		if strings.HasPrefix(response.Result.Objects[o], "heater_fan ") {
			heaterFans = append(heaterFans, strings.Replace(response.Result.Objects[o], "heater_fan ", "", 1))
		}
		// find filament_*_sensor
		if filamentSensorRegex.MatchString((response.Result.Objects[o])) {
			// The first element of the slice is the type of the sensor ("filament_{switch,motion}_sensor"),
			// and the second is the custom sensor's name itself
			filamentSensors = append(filamentSensors, []string{
				strings.TrimSpace(filamentSensorRegex.FindString(response.Result.Objects[o])),
				filamentSensorRegex.ReplaceAllString(response.Result.Objects[o], ""),
			})
		}
		// find heater_generic
		if strings.HasPrefix(response.Result.Objects[o], "heater_generic ") {
			genericHeaters = append(genericHeaters, strings.Replace(response.Result.Objects[o], "heater_generic ", "", 1))
		}
		// find tmc sensors
		if strings.HasPrefix(response.Result.Objects[o], "tmc") {
			// We need full name as the stepper type is part of the name
			tmcSensors = append(tmcSensors, response.Result.Objects[o])
		}
	}

	return &microcontrollers, &temperatureSensors, &temperatureFans, &temperatureProbes, &outputPins, &genericFans, &controllerFans, &heaterFans, &filamentSensors, &genericHeaters, &tmcSensors, nil
}

func (c Collector) fetchMoonrakerPrinterObjects() (*PrinterObjectResponse, error) {

	// Get the list of custom sensors if not already set. This saves fetching the full
	// list on every poll, but any new sensors will only be added is the exporter is restarted.
	// Double-checked locking: the check and HTTP fetch happen outside the mutex so that
	// distinct targets can initialize in parallel; only the map writes are serialized.
	customSensorsMu.Lock()
	_, ok := customTemperatureSensors[c.target]
	customSensorsMu.Unlock()

	if !ok {
		mcus, ts, tf, tp, op, gf, cf, hf, fs, gh, tmc, err := c.fetchCustomSensors()
		if err != nil {
			log.Error(err)
			return nil, err
		}
		log.Infof("Found custom sensors: %+v %+v %+v %+v %+v %+v %+v %+v %+v %+v", mcus, ts, tf, tp, op, gf, cf, hf, fs, gh)

		customSensorsMu.Lock()
		if _, ok := customTemperatureSensors[c.target]; !ok {
			customMicrocontrollers[c.target] = *mcus
			customTemperatureSensors[c.target] = *ts
			customTemperatureFans[c.target] = *tf
			customTemperatureProbes[c.target] = *tp
			customOutputPins[c.target] = *op
			customGenericFans[c.target] = *gf
			customControllerFans[c.target] = *cf
			customHeaterFans[c.target] = *hf
			customFilamentSensors[c.target] = *fs
			customGenericHeaters[c.target] = *gh
			customTmcSensors[c.target] = *tmc
		}
		customSensorsMu.Unlock()
	}

	mcuQuery := ""
	for mcu := range customMicrocontrollers[c.target] {
		if customMicrocontrollers[c.target][mcu] == "mcu" {
			mcuQuery += "&mcu=last_stats"
		} else {
			mcuQuery += "&mcu%20" + customMicrocontrollers[c.target][mcu] + "=last_stats"
		}
	}

	customSensorsQuery := ""
	for ts := range customTemperatureSensors[c.target] {
		customSensorsQuery += "&temperature_sensor%20" + customTemperatureSensors[c.target][ts]
	}
	for tf := range customTemperatureFans[c.target] {
		customSensorsQuery += "&temperature_fan%20" + customTemperatureFans[c.target][tf]
	}
	for tp := range customTemperatureProbes[c.target] {
		customSensorsQuery += "&temperature_probe%20" + customTemperatureProbes[c.target][tp]
	}
	for op := range customOutputPins[c.target] {
		customSensorsQuery += "&output_pin%20" + customOutputPins[c.target][op]
	}
	for gf := range customGenericFans[c.target] {
		customSensorsQuery += "&fan_generic%20" + customGenericFans[c.target][gf]
	}
	for cf := range customControllerFans[c.target] {
		customSensorsQuery += "&controller_fan%20" + customControllerFans[c.target][cf]
	}
	for hf := range customHeaterFans[c.target] {
		customSensorsQuery += "&heater_fan%20" + customHeaterFans[c.target][hf]
	}
	for fs := range customFilamentSensors[c.target] {
		customSensorsQuery += fmt.Sprintf("&%s%%20%s", customFilamentSensors[c.target][fs][0], customFilamentSensors[c.target][fs][1])
	}
	for gh := range customGenericHeaters[c.target] {
		customSensorsQuery += "&heater_generic%20" + customGenericHeaters[c.target][gh]
	}
	for tmc := range customTmcSensors[c.target] {
		customSensorsQuery += "&" + strings.ReplaceAll(customTmcSensors[c.target][tmc], " ", "%20")
	}

	urlPath := "/printer/objects/query" +
		"?" + gcodeMoveQuery +
		"&" + toolheadQuery +
		"&" + extruderQuery +
		"&" + heaterBedQuery +
		"&" + fanQuery +
		"&" + idleTimeoutQuery +
		"&" + virtualSdCardQuery +
		"&" + printStatsQuery +
		"&" + displayStatusQuery +
		"&" + heaterQuery +
		"&" + webhooksQuery +
		"&" + pauseResumeQuery +
		mcuQuery +
		customSensorsQuery

	var response PrinterObjectResponse
	if err := c.fetchFromMoonraker(urlPath, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

type QueryEndstopsResponse struct {
	Result map[string]string `json:"result"`
}

func (c Collector) fetchMoonrakerQueryEndstops() (map[string]string, error) {
	var response QueryEndstopsResponse
	if err := c.fetchFromMoonraker("/printer/query_endstops", &response); err != nil {
		return nil, err
	}
	return response.Result, nil
}

func (c Collector) collectQueryEndstops(ch chan<- prometheus.Metric) {
	log.Infof("Collecting query_endstops for %s", c.target)

	endstops, err := c.fetchMoonrakerQueryEndstops()
	if err != nil {
		log.Error(err)
		return
	}
	endstopLabels := []string{"endstop"}
	endstopDesc := prometheus.NewDesc("klipper_endstop_triggered", "Whether an endstop is triggered (1) or not (0).", endstopLabels, nil)
	for name, state := range endstops {
		ch <- prometheus.MustNewConstMetric(
			endstopDesc,
			prometheus.GaugeValue,
			boolToFloat64(state == "TRIGGERED"),
			GetValidLabelName(name))
	}
}

func (c Collector) collectPrinterObjects(ch chan<- prometheus.Metric) {
	log.Infof("Collecting printer_objects for %s", c.target)

	result, err := c.fetchMoonrakerPrinterObjects()
	if err != nil {
		log.Error(err)
		return
	}

	// gcode_move
	c.emitGauge(ch, "klipper_gcode_speed_factor", "Klipper gcode speed factor.", result.Result.Status.GcodeMove.SpeedFactor)
	c.emitGauge(ch, "klipper_gcode_speed", "Klipper gcode speed.", result.Result.Status.GcodeMove.Speed)
	c.emitGauge(ch, "klipper_gcode_extrude_factor", "Klipper gcode extrude factor.", result.Result.Status.GcodeMove.ExtrudeFactor)

	// gcode position
	if len(result.Result.Status.GcodeMove.GcodePosition) < 4 {
		log.Warn("Unexpected number of Gcode Position values, skipping gcode position metrics")
	} else {
		c.emitGauge(ch, "klipper_gcode_position_x", "Klipper gcode position X axis.", result.Result.Status.GcodeMove.GcodePosition[0])
		c.emitGauge(ch, "klipper_gcode_position_y", "Klipper gcode position Y axis.", result.Result.Status.GcodeMove.GcodePosition[1])
		c.emitGauge(ch, "klipper_gcode_position_z", "Klipper gcode position Z axis.", result.Result.Status.GcodeMove.GcodePosition[2])
		c.emitGauge(ch, "klipper_gcode_position_e", "Klipper gcode position for extruder.", result.Result.Status.GcodeMove.GcodePosition[3])
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
		sensorName := GetValidLabelName(mk)
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
	c.emitGauge(ch, "klipper_toolhead_print_time", "Klipper toolhead print time.", result.Result.Status.Toolhead.PrintTime)
	c.emitGauge(ch, "klipper_toolhead_estimated_print_time", "Klipper estimated print time.", result.Result.Status.Toolhead.EstimatedPrintTime)
	c.emitGauge(ch, "klipper_toolhead_max_velocity", "Klipper toolhead max velocity.", result.Result.Status.Toolhead.MaxVelocity)
	c.emitGauge(ch, "klipper_toolhead_max_accel", "Klipper toolhead max acceleration.", result.Result.Status.Toolhead.MaxAccel)
	c.emitGauge(ch, "klipper_toolhead_max_accel_to_decel", "Klipper toolhead max acceleration to deceleration.", result.Result.Status.Toolhead.MaxAccelToDecel)
	c.emitGauge(ch, "klipper_toolhead_square_corner_velocity", "Klipper toolhead square corner velocity.", result.Result.Status.Toolhead.SquareCornerVelocity)

	// toolhead homed axes
	for _, axis := range result.Result.Status.Toolhead.HomedAxes {
		emitStateInfoMetric(ch, "klipper_toolhead_homed_axes_info", "A homed axis on the toolhead.", "axis", string(axis))
	}
	c.emitCounter(ch, "klipper_toolhead_stalls_total", "Total number of toolhead stalls.", result.Result.Status.Toolhead.Stalls)

	// extruder
	c.emitGauge(ch, "klipper_extruder_temperature", "Klipper extruder temperature.", result.Result.Status.Extruder.Temperature)
	c.emitGauge(ch, "klipper_extruder_target", "Klipper extruder target.", result.Result.Status.Extruder.Target)
	c.emitGauge(ch, "klipper_extruder_power", "Klipper extruder power.", result.Result.Status.Extruder.Power)
	c.emitGauge(ch, "klipper_extruder_pressure_advance", "Klipper extruder pressure advance.", result.Result.Status.Extruder.PressureAdvance)
	c.emitGauge(ch, "klipper_extruder_smooth_time", "Klipper extruder smooth time.", result.Result.Status.Extruder.SmoothTime)

	// heater_bed
	c.emitGauge(ch, "klipper_heater_bed_temperature", "Klipper heater bed temperature.", result.Result.Status.HeaterBed.Temperature)
	c.emitGauge(ch, "klipper_heater_bed_target", "Klipper heater bed target.", result.Result.Status.HeaterBed.Target)
	c.emitGauge(ch, "klipper_heater_bed_power", "Klipper heater bed power.", result.Result.Status.HeaterBed.Power)

	// fan
	c.emitGauge(ch, "klipper_fan_speed", "Klipper fan speed.", result.Result.Status.Fan.Speed)
	if result.Result.Status.Fan.Rpm != nil {
		c.emitGauge(ch, "klipper_fan_rpm", "Klipper fan rpm.", *result.Result.Status.Fan.Rpm)
	}

	// idle_timeout
	c.emitCounter(ch, "klipper_printing_time", "The amount of time the printer has been in the Printing state.", result.Result.Status.IdleTimeout.PrintingTime)
	emitStateInfoMetric(ch, "klipper_idle_timeout_state_info", "The current idle timeout state of the printer.", "state", result.Result.Status.IdleTimeout.State)

	// virtual_sdcard
	c.emitCounter(ch, "klipper_print_file_progress", "The print progress reported as a percentage of the file read.", result.Result.Status.VirtualSdCard.Progress)
	c.emitCounter(ch, "klipper_print_file_position", "The current file position in bytes.", result.Result.Status.VirtualSdCard.FilePosition)
	c.emitGauge(ch, "klipper_sdcard_active", "Indicates whether the virtual SD card is actively being read (1) or not (0).", boolToFloat64(result.Result.Status.VirtualSdCard.IsActive))

	// print_stats
	c.emitCounter(ch, "klipper_print_total_duration", "The total time (in seconds) elapsed since a print has started.", result.Result.Status.PrintStats.TotalDuration)
	c.emitCounter(ch, "klipper_print_print_duration", "The total time spent printing (in seconds).", result.Result.Status.PrintStats.PrintDuration)
	c.emitCounter(ch, "klipper_print_filament_used", "The amount of filament used during the current print (in mm)..", result.Result.Status.PrintStats.FilamentUsed)

	// print state
	emitStateInfoMetric(ch, "klipper_print_state_info", "The current print state of the printer.", "state", result.Result.Status.PrintStats.State)
	if result.Result.Status.PrintStats.State != "" {
		c.emitGauge(ch, "klipper_printing", "Indicates whether the printer is currently printing (1) or not (0).", boolToFloat64(result.Result.Status.PrintStats.State == "printing"))
	}

	// webhooks
	emitStateInfoMetric(ch, "klipper_webhooks_state_info", "The current state of the Klipper webhooks server.", "state", result.Result.Status.Webhooks.State)

	// pause_resume
	c.emitGauge(ch, "klipper_pause_resume_is_paused", "Indicates whether the print is paused (1) or not (0).", boolToFloat64(result.Result.Status.PauseResume.IsPaused))

	// display_status
	c.emitCounter(ch, "klipper_print_gcode_progress", "The percentage of print progress, as reported by M73.", result.Result.Status.DisplayStatus.Progress)

	// temperature_sensor
	temperatureSensorLabels := []string{"sensor"}
	temperatureSensor := prometheus.NewDesc("klipper_temperature_sensor_temperature", "The temperature of the temperature sensor", temperatureSensorLabels, nil)
	temperatureSensorMinTemp := prometheus.NewDesc("klipper_temperature_sensor_measured_min_temp", "The measured minimum temperature of the temperature sensor", temperatureSensorLabels, nil)
	temperatureSensorMaxTemp := prometheus.NewDesc("klipper_temperature_sensor_measured_max_temp", "The measured maximum temperature of the temperature sensor", temperatureSensorLabels, nil)
	for sk, sv := range result.Result.Status.TemperatureSensors {
		sensorName := GetValidLabelName(sk)
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
	fanRpm := prometheus.NewDesc("klipper_temperature_fan_rpm", "The RPM of the temperature fan", fanLabels, nil)
	for fk, fv := range result.Result.Status.TemperatureFans {
		fanName := GetValidLabelName(fk)
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
		if fv.Rpm != nil {
			ch <- prometheus.MustNewConstMetric(
				fanRpm,
				prometheus.GaugeValue,
				*fv.Rpm,
				fanName)
		}
	}

	// temperature_probe
	temperatureProbeLabels := []string{"sensor"}
	temperatureProbe := prometheus.NewDesc("klipper_temperature_probe_temperature", "The temperature of the temperature probe", temperatureProbeLabels, nil)
	temperatureProbeMinTemp := prometheus.NewDesc("klipper_temperature_probe_measured_min_temp", "The measured minimum temperature of the temperature probe", temperatureProbeLabels, nil)
	temperatureProbeMaxTemp := prometheus.NewDesc("klipper_temperature_probe_measured_max_temp", "The measured maximum temperature of the temperature probe", temperatureProbeLabels, nil)
	temperatureProbeEstimatedExpansion := prometheus.NewDesc("klipper_temperature_probe_estimated_expansion", "The estimated of the temperature probe", temperatureProbeLabels, nil)
	for sk, sv := range result.Result.Status.TemperatureProbes {
		probeName := GetValidLabelName(sk)
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
		pinName := GetValidLabelName(k)
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
		fanName := GetValidLabelName(fk)
		ch <- prometheus.MustNewConstMetric(
			genericFanSpeed,
			prometheus.GaugeValue,
			fv.Speed,
			fanName)
		if fv.Rpm != nil {
			ch <- prometheus.MustNewConstMetric(
				genericFanRpm,
				prometheus.GaugeValue,
				*fv.Rpm,
				fanName)
		}
	}

	// controller_fan
	controllerFanLabels := []string{"fan"}
	controllerFanSpeed := prometheus.NewDesc("klipper_controller_fan_speed", "The speed of the controller fan", controllerFanLabels, nil)
	controllerFanRpm := prometheus.NewDesc("klipper_controller_fan_rpm", "The RPM of the controller fan", controllerFanLabels, nil)
	for fk, fv := range result.Result.Status.ControllerFans {
		fanName := GetValidLabelName(fk)
		ch <- prometheus.MustNewConstMetric(
			controllerFanSpeed,
			prometheus.GaugeValue,
			fv.Speed,
			fanName)
		if fv.Rpm != nil {
			ch <- prometheus.MustNewConstMetric(
				controllerFanRpm,
				prometheus.GaugeValue,
				*fv.Rpm,
				fanName)
		}
	}

	// heater_fan
	heaterFanLabels := []string{"fan"}
	heaterFanSpeed := prometheus.NewDesc("klipper_heater_fan_speed", "The speed of the heater fan", heaterFanLabels, nil)
	heaterFanRpm := prometheus.NewDesc("klipper_heater_fan_rpm", "The RPM of the heater fan", heaterFanLabels, nil)
	for fk, fv := range result.Result.Status.HeaterFans {
		fanName := GetValidLabelName(fk)
		ch <- prometheus.MustNewConstMetric(
			heaterFanSpeed,
			prometheus.GaugeValue,
			fv.Speed,
			fanName)
		if fv.Rpm != nil {
			ch <- prometheus.MustNewConstMetric(
				heaterFanRpm,
				prometheus.GaugeValue,
				*fv.Rpm,
				fanName)
		}
	}

	// filament_*_sensor
	filamentSensorLabels := []string{"sensor"}
	filamentSensorDetected := prometheus.NewDesc("klipper_filament_sensor_detected", "Whether filament presence is detected by the sensor", filamentSensorLabels, nil)
	filamentSensorEnabled := prometheus.NewDesc("klipper_filament_sensor_enabled", "Whether the filament sensor is enabled or not", filamentSensorLabels, nil)
	for k, v := range result.Result.Status.FilamentSensors {
		sensorName := GetValidLabelName(k)
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

	// heater_generic
	genericHeaterLabels := []string{"heater"}
	genericHeaterTemperature := prometheus.NewDesc("klipper_generic_heater_temperature", "The temperature of the generic heater", genericHeaterLabels, nil)
	genericHeaterTarget := prometheus.NewDesc("klipper_generic_heater_target", "The target temperature of the generic heater", genericHeaterLabels, nil)
	genericHeaterPower := prometheus.NewDesc("klipper_generic_heater_power", "The output power of the generic heater", genericHeaterLabels, nil)
	for name, heater := range result.Result.Status.GenericHeaters {
		heaterName := GetValidLabelName(name)
		ch <- prometheus.MustNewConstMetric(
			genericHeaterTemperature,
			prometheus.GaugeValue,
			heater.Temperature,
			heaterName)
		ch <- prometheus.MustNewConstMetric(
			genericHeaterTarget,
			prometheus.GaugeValue,
			heater.Target,
			heaterName)
		ch <- prometheus.MustNewConstMetric(
			genericHeaterPower,
			prometheus.GaugeValue,
			heater.Power,
			heaterName)
	}

	// tmc sensors
	tmcSensorLabels := []string{"sensor"}
	tmcTemperatureSensor := prometheus.NewDesc("klipper_tmc_sensor_temperature", "The temperature of the tmc driver", tmcSensorLabels, nil)
	tmcRunCurrentSensor := prometheus.NewDesc("klipper_tmc_sensor_run_current", "The run current of the tmc driver", tmcSensorLabels, nil)
	tmcEnabledSensor := prometheus.NewDesc("klipper_tmc_sensor_enabled", "Whether the tmc driver is enabled or not", tmcSensorLabels, nil)

	for sk, sv := range result.Result.Status.TmcSensors {
		sensorName := GetValidLabelName(strings.ReplaceAll(sk, " ", "_"))
		if sv.Temperature != nil {
			ch <- prometheus.MustNewConstMetric(
				tmcTemperatureSensor,
				prometheus.GaugeValue,
				*sv.Temperature,
				sensorName)
		}
		ch <- prometheus.MustNewConstMetric(
			tmcRunCurrentSensor,
			prometheus.GaugeValue,
			sv.RunCurrent,
			sensorName)

		ch <- prometheus.MustNewConstMetric(
			tmcEnabledSensor,
			prometheus.GaugeValue,
			boolToFloat64(sv.DrvStatus != nil),
			sensorName)
	}
}

package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#query-printer-object-status

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
)

type PrinterObjectResponse struct {
	Result struct {
		Status PrinterObjectStatus `json:"status"`
	} `json:"result"`
}

type PrinterObjectStatus struct {
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
	OutputPins         map[string]PrinterObjectOutputPin
	GenericFans        map[string]PrinterObjectFan
	ControllerFans     map[string]PrinterObjectFan
	FilamentSensors    map[string]PrinterObjectFilamentSensor
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
}

const toolheadQuery string = "toolhead=print_time,estimated_print_time,max_velocity,max_accel,max_accel_to_decel,square_corner_velocity"

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

type PrinterObjectFan struct {
	Speed float64 `json:"speed"`
	Rpm   float64 `json:"rpm"`
}

const fanQuery = "fan"

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
	TotalDuration float64 `json:"total_duration"`
	PrintDuration float64 `json:"print_duration"`
	FilamentUsed  float64 `json:"filament_used"`
}

const printStatsQuery = "print_stats=total_duration,print_duration,filament_used"

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
	Speed       float64 `mapstructure:"speed"`
	Temperature float64 `mapstructure:"temperature"`
	Target      float64 `mapstructure:"target"`
}

type PrinterObjectOutputPin struct {
	Value float64 `mapstructure:"value"`
}

type PrinterObjectFilamentSensor struct {
	Detected bool `mapstructure:"filament_detected"`
	Enabled  bool `mapstructure:"enabled"`
}

type _PrinterObjectStatus PrinterObjectStatus

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
		outputPins := make(map[string]PrinterObjectOutputPin)
		genericFans := make(map[string]PrinterObjectFan)
		controllerFans := make(map[string]PrinterObjectFan)
		filamentSensors := make(map[string]PrinterObjectFilamentSensor)
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
			if strings.HasPrefix(k, "output_pin") {
				key := strings.Replace(k, "output_pin ", "", 1)
				value := PrinterObjectOutputPin{}
				mapstructure.Decode(v, &value)
				outputPins[key] = value
			}
			if strings.HasPrefix(k, "fan_generic") {
				key := strings.Replace(k, "fan_generic ", "", 1)
				value := PrinterObjectFan{}
				mapstructure.Decode(v, &value)
				genericFans[key] = value
			}
			if strings.HasPrefix(k, "controller_fan") {
				key := strings.Replace(k, "controller_fan ", "", 1)
				value := PrinterObjectFan{}
				mapstructure.Decode(v, &value)
				controllerFans[key] = value
			}
			if filamentSensorRegex.MatchString(k) {
				key := strings.TrimSpace(filamentSensorRegex.ReplaceAllString(k, ""))
				value := PrinterObjectFilamentSensor{}
				mapstructure.Decode(v, &value)
				filamentSensors[key] = value
			}
		}
		f.Mcus = microcontrollers
		f.TemperatureSensors = temperatureSensors
		f.TemperatureFans = temperatureFans
		f.OutputPins = outputPins
		f.GenericFans = genericFans
		f.ControllerFans = controllerFans
		f.FilamentSensors = filamentSensors
	}
	return err
}

type PrinterObjectsList struct {
	Result struct {
		Objects []string `json:"objects"`
	} `json:"result"`
}

var (
	filamentSensorRegex      *regexp.Regexp        = regexp.MustCompile("^filament_(switch|motion)_sensor ")
	mcuRegex                 *regexp.Regexp        = regexp.MustCompile("^mcu(?P<label> [a-zA-Z0-9_]+)?")
	customMicrocontrollers   map[string][]string   = make(map[string][]string)
	customTemperatureSensors map[string][]string   = make(map[string][]string)
	customTemperatureFans    map[string][]string   = make(map[string][]string)
	customOutputPins         map[string][]string   = make(map[string][]string)
	customGenericFans        map[string][]string   = make(map[string][]string)
	customControllerFans     map[string][]string   = make(map[string][]string)
	customFilamentSensors    map[string][][]string = make(map[string][][]string)
)

// fetchCustomSensors queries klipper for the complete list and printer objects and
// returns the subset of `temperature_sensor`, `temperature_fan`, `output_pin`,
// `fan_generic`, `controller_fan`, and `filament_*_sensor` objects that have custom names.
func (c Collector) fetchCustomSensors(klipperHost string, apiKey string) (*[]string, *[]string, *[]string, *[]string, *[]string, *[]string, *[][]string, error) {
	var url = "http://" + klipperHost + "/printer/objects/list"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	var response PrinterObjectsList

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Error(err)
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	microcontrollers := []string{}
	temperatureSensors := []string{}
	temperatureFans := []string{}
	outputPins := []string{}
	genericFans := []string{}
	controllerFans := []string{}
	filamentSensors := [][]string{}
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
		// find filament_*_sensor
		if filamentSensorRegex.MatchString((response.Result.Objects[o])) {
			// The first element of the slice is the type of the sensor ("filament_{switch,motion}_sensor"),
			// and the second is the custom sensor's name itself
			filamentSensors = append(filamentSensors, []string{
				strings.TrimSpace(filamentSensorRegex.FindString(response.Result.Objects[o])),
				filamentSensorRegex.ReplaceAllString(response.Result.Objects[o], ""),
			})
		}
	}

	return &microcontrollers, &temperatureSensors, &temperatureFans, &outputPins, &genericFans, &controllerFans, &filamentSensors, nil
}

func (c Collector) fetchMoonrakerPrinterObjects(klipperHost string, apiKey string) (*PrinterObjectResponse, error) {

	// Get the list of custom sensors if not already set. This saves fetching the full
	// list on every poll, but any new sensors will only be added is the exporter is restarted.
	if _, ok := customTemperatureSensors[klipperHost]; ok {
		// already have custom sensors, skip
	} else {
		mcus, ts, tf, op, gf, cf, fs, err := c.fetchCustomSensors(klipperHost, apiKey)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		log.Infof("Found custom sensors: %+v %+v %+v %+v %+v %+v %+v", mcus, ts, tf, op, gf, cf, fs)
		customMicrocontrollers[klipperHost] = *mcus
		customTemperatureSensors[klipperHost] = *ts
		customTemperatureFans[klipperHost] = *tf
		customOutputPins[klipperHost] = *op
		customGenericFans[klipperHost] = *gf
		customControllerFans[klipperHost] = *cf
		customFilamentSensors[klipperHost] = *fs
	}

	mcuQuery := ""
	for mcu := range customMicrocontrollers[klipperHost] {
		if customMicrocontrollers[klipperHost][mcu] == "mcu" {
			mcuQuery += "&mcu=last_stats"
		} else {
			mcuQuery += "&mcu%20" + customMicrocontrollers[klipperHost][mcu] + "=last_stats"
		}
	}

	customSensorsQuery := ""
	for ts := range customTemperatureSensors[klipperHost] {
		customSensorsQuery += "&temperature_sensor%20" + customTemperatureSensors[klipperHost][ts]
	}
	for tf := range customTemperatureFans[klipperHost] {
		customSensorsQuery += "&temperature_fan%20" + customTemperatureFans[klipperHost][tf]
	}
	for op := range customOutputPins[klipperHost] {
		customSensorsQuery += "&output_pin%20" + customOutputPins[klipperHost][op]
	}
	for gf := range customGenericFans[klipperHost] {
		customSensorsQuery += "&fan_generic%20" + customGenericFans[klipperHost][gf]
	}
	for cf := range customControllerFans[klipperHost] {
		customSensorsQuery += "&controller_fan%20" + customControllerFans[klipperHost][cf]
	}
	for fs := range customFilamentSensors[klipperHost] {
		customSensorsQuery += fmt.Sprintf("&%s%%20%s", customFilamentSensors[klipperHost][fs][0], customFilamentSensors[klipperHost][fs][1])
	}

	var url = "http://" +
		klipperHost + "/printer/objects/query" +
		"?" + gcodeMoveQuery +
		"&" + toolheadQuery +
		"&" + extruderQuery +
		"&" + heaterBedQuery +
		"&" + fanQuery +
		"&" + idleTimeoutQuery +
		"&" + virtualSdCardQuery +
		"&" + printStatsQuery +
		"&" + displayStatusQuery +
		mcuQuery +
		customSensorsQuery

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

	log.Tracef("%+v", string(data))

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", res.StatusCode, res.Status)
	}

	var response PrinterObjectResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response data to %s. %s", reflect.TypeOf(response), err)
	}
	log.Tracef("%+v", response)

	return &response, nil
}

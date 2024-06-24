package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#query-printer-object-status

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"

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
	Mcu           PrinterObjectMcu           `json:"mcu"`
	// dynamic sensor attributes populated using custom unmarsaling
	TemperatureSensors map[string]PrinterObjectTemperatureSensor
	TemperatureFans    map[string]PrinterObjectTemperatureFan
	OutputPins         map[string]PrinterObjectOutputPin
	GenericFans        map[string]PrinterObjectFan
	ControllerFans     map[string]PrinterObjectFan
}

type PrinterObjectMcu struct {
	LastStats struct {
		McuAwake float64 `json:"mcu_awake"`
		// McuTaskAvg float64 `json:"mcu_task_avg"` // value returned in the format 1.5e-05
		// McuTaskStddev float64 `json:"mcu_task_stddev"` // value returned in the formate 1e-05
		BytesWrite      float64 `json:"bytes_write"`
		BytesRead       float64 `json:"bytes_read"`
		BytesRetransmit float64 `json:"bytes_retransmit"`
		BytesInvalid    float64 `json:"bytes_invalid"`
		SendSeq         float64 `json:"send_seq"`
		ReceiveSeq      float64 `json:"receive_seq"`
		RetransmitSeq   float64 `json:"retransmit_seq"`
		Srtt            float64 `json:"srtt"`
		Rttvar          float64 `json:"rttvar"`
		Rto             float64 `json:"rto"`
		ReadyBytes      float64 `json:"ready_bytes"`
		StalledBytes    float64 `json:"stalled_bytes"`
		Freq            float64 `json:"freq"`
	} `json:"last_stats"`
}

const mcuQuery string = "mcu=last_stats"

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
		temperatureSensors := make(map[string]PrinterObjectTemperatureSensor)
		temperatureFans := make(map[string]PrinterObjectTemperatureFan)
		outputPins := make(map[string]PrinterObjectOutputPin)
		genericFans := make(map[string]PrinterObjectFan)
		controllerFans := make(map[string]PrinterObjectFan)
		for k, v := range m {
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
		}
		f.TemperatureSensors = temperatureSensors
		f.TemperatureFans = temperatureFans
		f.OutputPins = outputPins
		f.GenericFans = genericFans
		f.ControllerFans = controllerFans
	}
	return err
}

type PrinterObjectsList struct {
	Result struct {
		Objects []string `json:"objects"`
	} `json:"result"`
}

var (
	customTemperatureSensors map[string][]string = make(map[string][]string)
	customTemperatureFans    map[string][]string = make(map[string][]string)
	customOutputPins         map[string][]string = make(map[string][]string)
	customGenericFans        map[string][]string = make(map[string][]string)
	customControllerFans     map[string][]string = make(map[string][]string)
)

// fetchCustomSensors queries klipper for the complete list and printer objects and
// returns the subset of `temperature_sensor`, `temperature_fan`, `output_pin`,
// `fan_generic`, and `controller_fan` objects that have custom names.
func (c Collector) fetchCustomSensors(klipperHost string, apiKey string) (*[]string, *[]string, *[]string, *[]string, *[]string, error) {
	var url = "http://" + klipperHost + "/printer/objects/list"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return nil, nil, nil, nil, nil, err
	}
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, nil, nil, nil, nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	var response PrinterObjectsList

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	temperatureSensors := []string{}
	temperatureFans := []string{}
	outputPins := []string{}
	genericFans := []string{}
	controllerFans := []string{}
	for o := range response.Result.Objects {
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
	}

	return &temperatureSensors, &temperatureFans, &outputPins, &genericFans, &controllerFans, nil
}

func (c Collector) fetchMoonrakerPrinterObjects(klipperHost string, apiKey string) (*PrinterObjectResponse, error) {

	// Get the list of custom sensors if not already set. This saves fetching the full
	// list on every poll, but any new sensors will only be added is the exporter is restarted.
	if _, ok := customTemperatureSensors[klipperHost]; ok {
		// already have custom sensors, skip
	} else {
		ts, tf, op, gf, cf, err := c.fetchCustomSensors(klipperHost, apiKey)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		log.Infof("Found custom sensors: %+v %+v %+v %+v %+v", ts, tf, op, gf, cf)
		customTemperatureSensors[klipperHost] = *ts
		customTemperatureFans[klipperHost] = *tf
		customOutputPins[klipperHost] = *op
		customGenericFans[klipperHost] = *gf
		customControllerFans[klipperHost] = *cf
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
		"&" + mcuQuery +
		customSensorsQuery

	log.Debug("Collecting metrics from " + url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Tracef("%+v", string(data))

	var response PrinterObjectResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Tracef("%+v", response)

	return &response, nil
}

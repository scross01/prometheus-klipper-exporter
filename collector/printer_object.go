package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#query-printer-object-status

import (
	"encoding/json"
	"io/ioutil"
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
	// dynamic sensor attributes populated using custom unmarsaling
	TemperatureSensors map[string]PrinterObjectTemperatureSensor
	TemperatureFans    map[string]PrinterObjectTemperatureFan
	OutputPins         map[string]PrinterObjectOutputPin
}

type PrinterObjectGcodeMove struct {
	SpeedFactor   float64 `json:"speed_factor"`
	Speed         float64 `json:"speed"`
	ExtrudeFactor float64 `json:"extrude_factor"`
}

const gcodeMoveQuery string = "gcode_move=speed_factor,speed,extrude_factor"

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
		}
		f.TemperatureSensors = temperatureSensors
		f.TemperatureFans = temperatureFans
		f.OutputPins = outputPins
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
)

// fetchCustomSensors queries klipper for the complete list and printer objects and
// returns the subset of `temperature_sensor`, `temperature_fan` and `output_pin`
// objects that have custom names.
func (c collector) fetchCustomSensors(klipperHost string, apiKey string) (*[]string, *[]string, *[]string, error) {
	var url = "http://" + klipperHost + "/printer/objects/list"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.Error(err)
		return nil, nil, nil, err
	}
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}
	res, err := client.Do(req)
	if err != nil {
		c.logger.Error(err)
		return nil, nil, nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.logger.Fatal(err)
		return nil, nil, nil, err
	}

	var response PrinterObjectsList

	err = json.Unmarshal(data, &response)
	if err != nil {
		c.logger.Fatal(err)
		return nil, nil, nil, err
	}

	temperatureSensors := []string{}
	temperatureFans := []string{}
	outputPins := []string{}
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
	}

	return &temperatureSensors, &temperatureFans, &outputPins, nil
}

func (c collector) fetchMoonrakerPrinterObjects(klipperHost string, apiKey string) (*PrinterObjectResponse, error) {

	// Get the list of custom sensors if not already set. This saves fetching the full
	// list on every poll, but any new sensors will only be added is the exporter is restarted.
	if _, ok := customTemperatureSensors[klipperHost]; ok {
		// already have custom sensors, skip
	} else {
		ts, tf, op, err := c.fetchCustomSensors(klipperHost, apiKey)
		if err != nil {
			c.logger.Error(err)
			return nil, err
		}
		c.logger.Infof("Found custom sensors: %+v %+v %+v", ts, tf, op)
		customTemperatureSensors[klipperHost] = *ts
		customTemperatureFans[klipperHost] = *tf
		customOutputPins[klipperHost] = *op
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
		customSensorsQuery

	c.logger.Debug("Collecting metrics from " + url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}
	res, err := client.Do(req)
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.logger.Fatal(err)
		return nil, err
	}

	c.logger.Tracef("%+v", string(data))

	var response PrinterObjectResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		c.logger.Fatal(err)
		return nil, err
	}
	c.logger.Tracef("%+v", response)

	return &response, nil
}

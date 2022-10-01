package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#query-printer-object-status

import (
	"encoding/json"
	"io/ioutil"

	"net/http"
)

type MoonrakerPrinterObjectResponse struct {
	Result struct {
		Status struct {
			GcodeMove PrinterObjectGcodeMove `json:"gcode_move"`
			Toolhead  PrinterObjectToolhead  `json:"toolhead"`
			Extruder  PrinterObjectExtruder  `json:"extruder"`
			HeaterBed PrinterObjectHeaterBed `json:"heater_bed`
			Fan       PrinterObjectFan       `json:"fan"`
		} `json:"status"`
	} `json:"result"`
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
	Rpm   int32   `json:"rpm"`
}

const fanQuery = "fan"

func (c collector) fetchMoonrakerPrinterObjects(klipperHost string) (*MoonrakerPrinterObjectResponse, error) {
	var procStatsUrl = "http://" +
		klipperHost + "/printer/objects/query" +
		"?" + gcodeMoveQuery +
		"&" + toolheadQuery +
		"&" + extruderQuery +
		"&" + heaterBedQuery +
		"&" + fanQuery
	c.logger.Debug("Collecting metrics from " + procStatsUrl)
	res, err := http.Get(procStatsUrl)

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

	var response MoonrakerPrinterObjectResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		c.logger.Fatal(err)
		return nil, err
	}

	return &response, nil
}

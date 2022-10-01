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
			GcodeMove     PrinterObjectGcodeMove     `json:"gcode_move"`
			Toolhead      PrinterObjectToolhead      `json:"toolhead"`
			Extruder      PrinterObjectExtruder      `json:"extruder"`
			HeaterBed     PrinterObjectHeaterBed     `json:"heater_bed"`
			Fan           PrinterObjectFan           `json:"fan"`
			IdleTimeout   PrinterObjectIdleTimeout   `json:"idle_timeout"`
			VirtualSdCard PrinterObjectVirtualSdCard `json:"virtual_sdcard"`
			PrintStats    PrinterObjectPrintStats    `json:"print_stats"`
			DisplayStatus PrinterObjectDisplayStatus `json:"display_status"`
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

const DisplayStatusQuery = "display_status"

func (c collector) fetchMoonrakerPrinterObjects(klipperHost string) (*MoonrakerPrinterObjectResponse, error) {
	var procStatsUrl = "http://" +
		klipperHost + "/printer/objects/query" +
		"?" + gcodeMoveQuery +
		"&" + toolheadQuery +
		"&" + extruderQuery +
		"&" + heaterBedQuery +
		"&" + fanQuery +
		"&" + idleTimeoutQuery +
		"&" + virtualSdCardQuery +
		"&" + printStatsQuery +
		"&" + DisplayStatusQuery
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

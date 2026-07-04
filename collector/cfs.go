package collector

// CFS (Creality Filament System) collector
// https://moonraker.readthedocs.io/en/latest/external_api/printer/
//
// Creality K2-class printers do not populate Happy-Hare's `mmu` object. Instead the
// filament system state lives in native Moonraker objects:
//   - `box`           : the CFS unit(s) and per-slot state (active slot, temp, humidity)
//   - `filament_rack`  : the filament currently loaded at the toolhead
//
// The payload mixes string-encoded numbers and uses "None" / "-1" sentinels on
// disconnected units. We avoid all custom unmarshalling by only declaring the fields
// we emit (encoding/json ignores the rest) and skipping units whose state == "None".

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// CFS response structures
type CFSResponse struct {
	Result struct {
		EventTime float64   `json:"eventtime"`
		Status    CFSStatus `json:"status"`
	} `json:"result"`
}

type CFSStatus struct {
	Box          BoxObject          `json:"box"`
	FilamentRack FilamentRackObject `json:"filament_rack"`
}

type BoxObject struct {
	Filament      int               `json:"filament"`
	State         string            `json:"state"`
	AutoRefill    int               `json:"auto_refill"`
	Enable        int               `json:"enable"`
	FilamentUseup int               `json:"filament_useup"`
	Map           map[string]string `json:"map"`
	T1            BoxUnit           `json:"T1"`
	T2            BoxUnit           `json:"T2"`
	T3            BoxUnit           `json:"T3"`
	T4            BoxUnit           `json:"T4"`
	// Intentionally omitted (mixed types, deferred): same_material, plus per-unit
	// measuring_wheel / uuid / change_color_num / filament_detected / slot_rfid_scrap
}

type BoxUnit struct {
	State          string   `json:"state"`
	Filament       string   `json:"filament"`         // active slot letter, or "None"
	Temperature    string   `json:"temperature"`      // string-encoded number
	DryAndHumidity string   `json:"dry_and_humidity"` // string-encoded number
	Version        string   `json:"version"`
	Sn             string   `json:"sn"`
	Mode           string   `json:"mode"`
	Vender         []string `json:"vender"`        // per slot A..D (sic: Creality spelling)
	RemainLen      []string `json:"remain_len"`    // per slot, string-encoded
	ColorValue     []string `json:"color_value"`   // per slot
	MaterialType   []string `json:"material_type"` // per slot, code
}

// FilamentRackObject describes the filament currently loaded at the toolhead. On this
// hardware vender/color_value/material_type are "-1" placeholders; the live values are
// in the remain_material_* fields.
type FilamentRackObject struct {
	RemainMaterialColor    string  `json:"remain_material_color"`
	RemainMaterialType     string  `json:"remain_material_type"`
	RemainMaterialVelocity float64 `json:"remain_material_velocity"`
}

// slotLetterToIndex maps a CFS slot letter to its array index (A=0..D=3), returning
// -1 for anything else (including the "None" sentinel on empty units).
func slotLetterToIndex(s string) int {
	switch s {
	case "A":
		return 0
	case "B":
		return 1
	case "C":
		return 2
	case "D":
		return 3
	default:
		return -1
	}
}

// parseCFSFloat parses a string-encoded number, treating "" and "None" as absent.
func parseCFSFloat(s string) (float64, bool) {
	if s == "" || s == "None" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// Fetch CFS data from Moonraker
func (c Collector) fetchCFSData() (*CFSResponse, error) {
	var response CFSResponse
	if err := c.fetchFromMoonraker("/printer/objects/query?box&filament_rack", &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c Collector) collectCFS(ch chan<- prometheus.Metric) {
	result, err := c.fetchCFSData()
	if err != nil {
		log.Errorf("Failed to fetch CFS data: %v", err)
		return
	}

	box := result.Result.Status.Box
	rack := result.Result.Status.FilamentRack

	// === Box-level metrics ===
	c.emitGauge(ch, "klipper_cfs_enabled", "CFS enabled state", boolToFloat64(box.Enable == 1))
	c.emitGauge(ch, "klipper_cfs_auto_refill_enabled", "CFS auto-refill enabled", boolToFloat64(box.AutoRefill == 1))
	c.emitGauge(ch, "klipper_cfs_filament_useup", "CFS filament used up flag", boolToFloat64(box.FilamentUseup == 1))
	emitStateInfoMetric(ch, "klipper_cfs_state_info", "CFS connection state", "state", box.State)
	// NOTE: box.filament semantics are unconfirmed (active unit number? loaded count?).
	c.emitGauge(ch, "klipper_cfs_active_unit", "CFS box.filament value (semantics unconfirmed: likely active unit number)", float64(box.Filament))

	// === Per-unit and per-slot metrics (skip disconnected units) ===
	units := []struct {
		name string
		unit BoxUnit
	}{
		{"T1", box.T1},
		{"T2", box.T2},
		{"T3", box.T3},
		{"T4", box.T4},
	}

	unitLabels := []string{"unit"}
	activeSlotDesc := prometheus.NewDesc("klipper_cfs_active_slot", "Active slot index within the unit (A=0..D=3, -1 if none)", unitLabels, nil)
	activeSlotInfoDesc := prometheus.NewDesc("klipper_cfs_active_slot_info", "Active slot details (always 1)", []string{"unit", "slot", "material", "color"}, nil)
	unitTempDesc := prometheus.NewDesc("klipper_cfs_unit_temperature_celsius", "CFS unit temperature in celsius", unitLabels, nil)
	unitHumidityDesc := prometheus.NewDesc("klipper_cfs_unit_humidity_percent", "CFS unit relative humidity percent (assumed %RH)", unitLabels, nil)
	unitStateDesc := prometheus.NewDesc("klipper_cfs_unit_state_info", "CFS unit connection state (always 1)", []string{"unit", "state"}, nil)
	unitInfoDesc := prometheus.NewDesc("klipper_cfs_unit_info", "CFS unit hardware information (always 1)", []string{"unit", "version", "sn", "mode"}, nil)
	slotInfoDesc := prometheus.NewDesc("klipper_cfs_slot_info", "CFS slot details (always 1)", []string{"unit", "slot", "material", "color", "vendor"}, nil)
	slotRemainingDesc := prometheus.NewDesc("klipper_cfs_slot_remaining", "CFS slot remaining filament (units unclear: percent or mm)", []string{"unit", "slot"}, nil)

	slotLetters := []string{"A", "B", "C", "D"}

	for _, u := range units {
		if u.unit.State == "None" {
			continue
		}

		// Unit state and hardware info
		ch <- prometheus.MustNewConstMetric(unitStateDesc, prometheus.GaugeValue, 1, u.name, u.unit.State)
		ch <- prometheus.MustNewConstMetric(unitInfoDesc, prometheus.GaugeValue, 1, u.name, u.unit.Version, u.unit.Sn, u.unit.Mode)

		if temp, ok := parseCFSFloat(u.unit.Temperature); ok {
			ch <- prometheus.MustNewConstMetric(unitTempDesc, prometheus.GaugeValue, temp, u.name)
		}
		if humidity, ok := parseCFSFloat(u.unit.DryAndHumidity); ok {
			ch <- prometheus.MustNewConstMetric(unitHumidityDesc, prometheus.GaugeValue, humidity, u.name)
		}

		// Active slot (the deliverable)
		idx := slotLetterToIndex(u.unit.Filament)
		ch <- prometheus.MustNewConstMetric(activeSlotDesc, prometheus.GaugeValue, float64(idx), u.name)
		if idx >= 0 {
			material := ""
			color := ""
			if idx < len(u.unit.MaterialType) {
				material = u.unit.MaterialType[idx]
			}
			if idx < len(u.unit.ColorValue) {
				color = u.unit.ColorValue[idx]
			}
			ch <- prometheus.MustNewConstMetric(activeSlotInfoDesc, prometheus.GaugeValue, 1, u.name, u.unit.Filament, material, color)
		}

		// Per-slot metrics
		for i, letter := range slotLetters {
			material := ""
			color := ""
			vendor := ""
			if i < len(u.unit.MaterialType) {
				material = u.unit.MaterialType[i]
			}
			if i < len(u.unit.ColorValue) {
				color = u.unit.ColorValue[i]
			}
			if i < len(u.unit.Vender) {
				vendor = u.unit.Vender[i]
			}
			ch <- prometheus.MustNewConstMetric(slotInfoDesc, prometheus.GaugeValue, 1, u.name, letter, material, color, vendor)

			if i < len(u.unit.RemainLen) {
				if remain, ok := parseCFSFloat(u.unit.RemainLen[i]); ok {
					ch <- prometheus.MustNewConstMetric(slotRemainingDesc, prometheus.GaugeValue, remain, u.name, letter)
				}
			}
		}
	}

	// === Filament rack (loaded at toolhead) ===
	emitStateInfoMetric2(ch, "klipper_cfs_rack_loaded_info", "Filament currently loaded at the toolhead (always 1)",
		"material", rack.RemainMaterialType, "color", rack.RemainMaterialColor)
	c.emitGauge(ch, "klipper_cfs_rack_velocity", "Loaded filament velocity (units unclear, likely mm/min)", rack.RemainMaterialVelocity)
}

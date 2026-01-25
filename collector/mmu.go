package collector

// MMU (Multi-Material Unit) collector for Happy Hare
// https://github.com/moggieuk/Happy-Hare
// https://github.com/moggieuk/Happy-Hare/wiki/Printer-Variables
// https://moonraker.readthedocs.io/en/latest/external_api/printer/
// Unfortunately the MMU metrics are not well documented, so some interpretation is needed.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// MMU Response structures
type MMUResponse struct {
	Result struct {
		EventTime float64   `json:"eventtime"`
		Status    MMUStatus `json:"status"`
	} `json:"result"`
}

type MMUStatus struct {
	MMU        MMUObject        `json:"mmu"`
	MMUEncoder MMUEncoderObject `json:"mmu_encoder mmu_encoder"`
	MMUMachine MMUMachineObject `json:"mmu_machine"`
}

type MMUObject struct {
	// Basic state
	Enabled   bool `json:"enabled"`
	NumGates  int  `json:"num_gates"`
	IsHomed   bool `json:"is_homed"`
	HasBypass bool `json:"has_bypass"`
	Unit      int  `json:"unit"`
	Tool      int  `json:"tool"`
	Gate      int  `json:"gate"`

	// Print state
	PrintState     string `json:"print_state"`
	Operation      string `json:"operation"`
	Action         string `json:"action"`
	Runout         bool   `json:"runout"`
	ReasonForPause string `json:"reason_for_pause"`

	// Filament state
	Filament          string  `json:"filament"`
	FilamentPosition  float64 `json:"filament_position"`
	FilamentPos       int     `json:"filament_pos"`
	FilamentDirection int     `json:"filament_direction"`

	// Toolchange info
	NumToolchanges        int     `json:"num_toolchanges"`
	LastTool              int     `json:"last_tool"`
	NextTool              int     `json:"next_tool"`
	ToolchangePurgeVolume float64 `json:"toolchange_purge_volume"`
	LastToolchange        string  `json:"last_toolchange"`

	// Active filament
	ActiveFilament MMUActiveFilament `json:"active_filament"`

	// Detection settings
	ClogDetectionEnabled int `json:"clog_detection_enabled"`
	EndlessSpoolEnabled  int `json:"endless_spool_enabled"`
	SyncFeedbackEnabled  int `json:"sync_feedback_enabled"`

	// Sync drive
	SyncDrive         bool   `json:"sync_drive"`
	SyncFeedbackState string `json:"sync_feedback_state"`

	// Servo
	Servo string `json:"servo"`

	// Bowden
	BowdenProgress int `json:"bowden_progress"`

	// Spoolman
	SpoolmanSupport string `json:"spoolman_support"`
	PendingSpoolId  int    `json:"pending_spool_id"`

	// Espooler
	EspoolerActive string `json:"espooler_active"`

	// Per-gate arrays
	TTGMap             []int    `json:"ttg_map"`
	EndlessSpoolGroups []int    `json:"endless_spool_groups"`
	GateStatus         []int    `json:"gate_status"`
	GateFilamentName   []string `json:"gate_filament_name"`
	GateMaterial       []string `json:"gate_material"`
	GateColor          []string `json:"gate_color"`
	GateTemperature    []int    `json:"gate_temperature"`
	GateSpoolId        []int    `json:"gate_spool_id"`
	GateSpeedOverride  []int    `json:"gate_speed_override"`

	// Tool multipliers
	ToolExtrusionMultipliers []float64 `json:"tool_extrusion_multipliers"`
	ToolSpeedMultipliers     []float64 `json:"tool_speed_multipliers"`

	// Sensors
	Sensors map[string]interface{} `json:"sensors"`

	// Encoder (nested)
	Encoder MMUEncoderData `json:"encoder"`

	// Slicer tool map
	SlicerToolMap MMUSlicerToolMap `json:"slicer_tool_map"`
}

type MMUActiveFilament struct {
	FilamentName string `json:"filament_name"`
	Material     string `json:"material"`
	Color        string `json:"color"`
	SpoolId      int    `json:"spool_id"`
	Temperature  int    `json:"temperature"`
}

type MMUEncoderData struct {
	EncoderPos      float64 `json:"encoder_pos"`
	DetectionLength float64 `json:"detection_length"`
	MinHeadroom     float64 `json:"min_headroom"`
	Headroom        float64 `json:"headroom"`
	DesiredHeadroom float64 `json:"desired_headroom"`
	DetectionMode   int     `json:"detection_mode"`
	Enabled         bool    `json:"enabled"`
	FlowRate        int     `json:"flow_rate"`
}

type MMUEncoderObject struct {
	EncoderPos      float64 `json:"encoder_pos"`
	DetectionLength float64 `json:"detection_length"`
	MinHeadroom     float64 `json:"min_headroom"`
	Headroom        float64 `json:"headroom"`
	DesiredHeadroom float64 `json:"desired_headroom"`
	DetectionMode   int     `json:"detection_mode"`
	Enabled         bool    `json:"enabled"`
	FlowRate        int     `json:"flow_rate"`
}

type MMUMachineObject struct {
	NumUnits int            `json:"num_units"`
	Unit0    MMUMachineUnit `json:"unit_0"`
}

type MMUMachineUnit struct {
	Name         string `json:"name"`
	Vendor       string `json:"vendor"`
	Version      string `json:"version"`
	NumGates     int    `json:"num_gates"`
	FirstGate    int    `json:"first_gate"`
	SelectorType string `json:"selector_type"`
	HasBypass    bool   `json:"has_bypass"`
}

type MMUSlicerToolMap struct {
	ReferencedTools  []int `json:"referenced_tools"`
	InitialTool      int   `json:"initial_tool"`
	TotalToolchanges int   `json:"total_toolchanges"`
	SkipAutomap      bool  `json:"skip_automap"`
}

// Fetch MMU data from Moonraker
func (c Collector) fetchMMUData(klipperHost string, apiKey string) (*MMUResponse, error) {
	url := "http://" + klipperHost + "/printer/objects/query?mmu&mmu_encoder%20mmu_encoder&mmu_machine"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP request for %s: %w", url, err)
	}
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to complete HTTP request: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	log.Tracef("MMU Response: %s", string(data))

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", res.StatusCode, res.Status)
	}

	var response MMUResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response to %s: %w", reflect.TypeOf(response), err)
	}

	return &response, nil
}

// Fetch MMU pre-gate sensors
func (c Collector) fetchMMUPreGateSensors(klipperHost string, apiKey string, numGates int) ([]bool, []bool, error) {
	// Build query for all pre-gate sensors
	query := ""
	for i := 0; i < numGates; i++ {
		if i > 0 {
			query += "&"
		}
		query += fmt.Sprintf("filament_switch_sensor%%20mmu_pre_gate_%d", i)
	}

	url := "http://" + klipperHost + "/printer/objects/query?" + query

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create HTTP request: %w", err)
	}
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to complete HTTP request: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, nil, fmt.Errorf("unable to unmarshal response: %w", err)
	}

	detected := make([]bool, numGates)
	enabled := make([]bool, numGates)

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		if status, ok := resultData["status"].(map[string]interface{}); ok {
			for i := 0; i < numGates; i++ {
				sensorKey := fmt.Sprintf("filament_switch_sensor mmu_pre_gate_%d", i)
				if sensor, ok := status[sensorKey].(map[string]interface{}); ok {
					if d, ok := sensor["filament_detected"].(bool); ok {
						detected[i] = d
					}
					if e, ok := sensor["enabled"].(bool); ok {
						enabled[i] = e
					}
				}
			}
		}
	}

	return detected, enabled, nil
}

func (c Collector) collectMMU(ch chan<- prometheus.Metric) {
	log.Infof("Collecting mmu for %s", c.target)

	result, err := c.fetchMMUData(c.target, c.apiKey)
	if err != nil {
		log.Errorf("Failed to fetch MMU data: %v", err)
		return
	}

	mmu := result.Result.Status.MMU
	machine := result.Result.Status.MMUMachine

	// === Basic State Metrics ===
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_enabled", "MMU enabled state", nil, nil),
		prometheus.GaugeValue,
		boolToFloat64(mmu.Enabled))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_homed", "MMU homed state", nil, nil),
		prometheus.GaugeValue,
		boolToFloat64(mmu.IsHomed))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_num_gates", "Number of MMU gates", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.NumGates))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_has_bypass", "MMU has bypass gate", nil, nil),
		prometheus.GaugeValue,
		boolToFloat64(mmu.HasBypass))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_current_unit", "Current MMU unit", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.Unit))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_current_tool", "Current tool (-1=unknown, -2=bypass)", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.Tool))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_current_gate", "Current gate", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.Gate))

	// === Print State ===
	if mmu.PrintState != "" {
		printStateDesc := prometheus.NewDesc("klipper_mmu_print_state_info", "MMU print state", []string{"state"}, nil)
		ch <- prometheus.MustNewConstMetric(printStateDesc, prometheus.GaugeValue, 1.0, mmu.PrintState)
	}

	// === Action State ===
	if mmu.Action != "" {
		actionDesc := prometheus.NewDesc("klipper_mmu_action_info", "MMU current action", []string{"action"}, nil)
		ch <- prometheus.MustNewConstMetric(actionDesc, prometheus.GaugeValue, 1.0, mmu.Action)
	}

	// === Operation State ===
	if mmu.Operation != "" {
		operationDesc := prometheus.NewDesc("klipper_mmu_operation_info", "MMU current operation", []string{"operation"}, nil)
		ch <- prometheus.MustNewConstMetric(operationDesc, prometheus.GaugeValue, 1.0, mmu.Operation)
	}

	// === Filament State ===
	filamentLoaded := 0.0
	if mmu.Filament == "Loaded" {
		filamentLoaded = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_filament_loaded", "Filament loaded state", nil, nil),
		prometheus.GaugeValue,
		filamentLoaded)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_filament_position_mm", "Filament position in mm", nil, nil),
		prometheus.GaugeValue,
		mmu.FilamentPosition)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_filament_pos_state", "Filament position state machine value", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.FilamentPos))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_filament_direction", "Filament direction (1=load, -1=unload)", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.FilamentDirection))

	// === Toolchange Metrics ===
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_toolchanges_total", "Total toolchanges in current print", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.NumToolchanges))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_last_tool", "Last tool used", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.LastTool))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_next_tool", "Next tool during toolchange", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.NextTool))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_toolchange_purge_volume_mm3", "Suggested purge volume in mm³", nil, nil),
		prometheus.GaugeValue,
		mmu.ToolchangePurgeVolume)

	// === Runout ===
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_runout", "Runout detected", nil, nil),
		prometheus.GaugeValue,
		boolToFloat64(mmu.Runout))

	// === Detection Settings ===
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_clog_detection_mode", "Clog detection mode (0=off, 1=manual, 2=auto)", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.ClogDetectionEnabled))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_endless_spool_enabled", "Endless spool enabled (0=off, 1=enabled, 2=pre-gate)", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.EndlessSpoolEnabled))

	// === Sync Drive ===
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_sync_drive_enabled", "Gear stepper synced to extruder", nil, nil),
		prometheus.GaugeValue,
		boolToFloat64(mmu.SyncDrive))

	// Sync feedback state
	if mmu.SyncFeedbackState != "" {
		syncStateDesc := prometheus.NewDesc("klipper_mmu_sync_feedback_state_info", "Sync feedback state", []string{"state"}, nil)
		ch <- prometheus.MustNewConstMetric(syncStateDesc, prometheus.GaugeValue, 1.0, mmu.SyncFeedbackState)
	}

	// === Servo Position ===
	if mmu.Servo != "" {
		servoDesc := prometheus.NewDesc("klipper_mmu_servo_position_info", "Servo position", []string{"position"}, nil)
		ch <- prometheus.MustNewConstMetric(servoDesc, prometheus.GaugeValue, 1.0, mmu.Servo)
	}

	// === Bowden Progress ===
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_bowden_progress_percent", "Bowden move progress (-1 if not active)", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.BowdenProgress))

	// === Encoder Metrics ===
	encoder := mmu.Encoder
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_encoder_position_mm", "Encoder position in mm", nil, nil),
		prometheus.GaugeValue,
		encoder.EncoderPos)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_encoder_detection_length_mm", "Clog detection length in mm", nil, nil),
		prometheus.GaugeValue,
		encoder.DetectionLength)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_encoder_headroom_mm", "Current clog detection headroom in mm", nil, nil),
		prometheus.GaugeValue,
		encoder.Headroom)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_encoder_min_headroom_mm", "Minimum headroom recorded in mm", nil, nil),
		prometheus.GaugeValue,
		encoder.MinHeadroom)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_encoder_desired_headroom_mm", "Desired headroom in mm", nil, nil),
		prometheus.GaugeValue,
		encoder.DesiredHeadroom)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_encoder_flow_rate_percent", "Encoder flow rate percent", nil, nil),
		prometheus.GaugeValue,
		float64(encoder.FlowRate))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_encoder_enabled", "Encoder enabled for clog detection", nil, nil),
		prometheus.GaugeValue,
		boolToFloat64(encoder.Enabled))

	// === Per-Gate Metrics ===
	gateLabels := []string{"gate"}
	gateStatusDesc := prometheus.NewDesc("klipper_mmu_gate_status", "Gate status (-1=unknown, 0=empty, 1=available, 2=buffered)", gateLabels, nil)
	gateTemperatureDesc := prometheus.NewDesc("klipper_mmu_gate_temperature", "Gate filament temperature", gateLabels, nil)
	gateSpeedOverrideDesc := prometheus.NewDesc("klipper_mmu_gate_speed_override_percent", "Gate speed override percent", gateLabels, nil)
	gateTTGMapDesc := prometheus.NewDesc("klipper_mmu_gate_ttg_map", "Tool-to-gate mapping value", gateLabels, nil)
	gateEndlessSpoolGroupDesc := prometheus.NewDesc("klipper_mmu_gate_endless_spool_group", "Endless spool group", gateLabels, nil)
	gateSpoolIdDesc := prometheus.NewDesc("klipper_mmu_gate_spool_id", "Spoolman spool ID (-1 if not set)", gateLabels, nil)

	for i := 0; i < mmu.NumGates; i++ {
		gateStr := strconv.Itoa(i)

		if i < len(mmu.GateStatus) {
			ch <- prometheus.MustNewConstMetric(gateStatusDesc, prometheus.GaugeValue, float64(mmu.GateStatus[i]), gateStr)
		}
		if i < len(mmu.GateTemperature) {
			ch <- prometheus.MustNewConstMetric(gateTemperatureDesc, prometheus.GaugeValue, float64(mmu.GateTemperature[i]), gateStr)
		}
		if i < len(mmu.GateSpeedOverride) {
			ch <- prometheus.MustNewConstMetric(gateSpeedOverrideDesc, prometheus.GaugeValue, float64(mmu.GateSpeedOverride[i]), gateStr)
		}
		if i < len(mmu.TTGMap) {
			ch <- prometheus.MustNewConstMetric(gateTTGMapDesc, prometheus.GaugeValue, float64(mmu.TTGMap[i]), gateStr)
		}
		if i < len(mmu.EndlessSpoolGroups) {
			ch <- prometheus.MustNewConstMetric(gateEndlessSpoolGroupDesc, prometheus.GaugeValue, float64(mmu.EndlessSpoolGroups[i]), gateStr)
		}
		if i < len(mmu.GateSpoolId) {
			ch <- prometheus.MustNewConstMetric(gateSpoolIdDesc, prometheus.GaugeValue, float64(mmu.GateSpoolId[i]), gateStr)
		}
	}

	// === Gate Info (with material/color labels) ===
	gateInfoLabels := []string{"gate", "material", "color", "filament_name"}
	gateInfoDesc := prometheus.NewDesc("klipper_mmu_gate_info", "Gate information (always 1)", gateInfoLabels, nil)
	for i := 0; i < mmu.NumGates; i++ {
		gateStr := strconv.Itoa(i)
		material := ""
		color := ""
		name := ""
		if i < len(mmu.GateMaterial) {
			material = mmu.GateMaterial[i]
		}
		if i < len(mmu.GateColor) {
			color = mmu.GateColor[i]
		}
		if i < len(mmu.GateFilamentName) {
			name = mmu.GateFilamentName[i]
		}
		ch <- prometheus.MustNewConstMetric(gateInfoDesc, prometheus.GaugeValue, 1, gateStr, material, color, name)
	}

	// === Tool Multipliers ===
	toolLabels := []string{"tool"}
	extrusionMultiplierDesc := prometheus.NewDesc("klipper_mmu_tool_extrusion_multiplier", "Tool extrusion multiplier (M221)", toolLabels, nil)
	speedMultiplierDesc := prometheus.NewDesc("klipper_mmu_tool_speed_multiplier", "Tool speed multiplier (M220)", toolLabels, nil)

	for i := 0; i < mmu.NumGates; i++ {
		toolStr := strconv.Itoa(i)
		if i < len(mmu.ToolExtrusionMultipliers) {
			ch <- prometheus.MustNewConstMetric(extrusionMultiplierDesc, prometheus.GaugeValue, mmu.ToolExtrusionMultipliers[i], toolStr)
		}
		if i < len(mmu.ToolSpeedMultipliers) {
			ch <- prometheus.MustNewConstMetric(speedMultiplierDesc, prometheus.GaugeValue, mmu.ToolSpeedMultipliers[i], toolStr)
		}
	}

	// === Pre-Gate Sensors ===
	detected, enabled, err := c.fetchMMUPreGateSensors(c.target, c.apiKey, mmu.NumGates)
	if err != nil {
		log.Warnf("Failed to fetch pre-gate sensors: %v", err)
	} else {
		preGateDetectedDesc := prometheus.NewDesc("klipper_mmu_pre_gate_sensor_detected", "Pre-gate sensor filament detected", gateLabels, nil)
		preGateEnabledDesc := prometheus.NewDesc("klipper_mmu_pre_gate_sensor_enabled", "Pre-gate sensor enabled", gateLabels, nil)

		for i := 0; i < mmu.NumGates; i++ {
			gateStr := strconv.Itoa(i)
			ch <- prometheus.MustNewConstMetric(preGateDetectedDesc, prometheus.GaugeValue, boolToFloat64(detected[i]), gateStr)
			ch <- prometheus.MustNewConstMetric(preGateEnabledDesc, prometheus.GaugeValue, boolToFloat64(enabled[i]), gateStr)
		}
	}

	// === Slicer Tool Map Info ===
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_slicer_total_toolchanges", "Total toolchanges expected from slicer", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.SlicerToolMap.TotalToolchanges))

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_slicer_initial_tool", "Initial tool from slicer", nil, nil),
		prometheus.GaugeValue,
		float64(mmu.SlicerToolMap.InitialTool))

	// === Machine Info ===
	machineInfoLabels := []string{"name", "vendor", "version", "selector_type"}
	machineInfoDesc := prometheus.NewDesc("klipper_mmu_machine_info", "MMU machine information (always 1)", machineInfoLabels, nil)
	ch <- prometheus.MustNewConstMetric(
		machineInfoDesc,
		prometheus.GaugeValue,
		1,
		machine.Unit0.Name,
		machine.Unit0.Vendor,
		machine.Unit0.Version,
		machine.Unit0.SelectorType)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("klipper_mmu_num_units", "Number of MMU units", nil, nil),
		prometheus.GaugeValue,
		float64(machine.NumUnits))

	// === Active Filament Info ===
	if mmu.ActiveFilament.FilamentName != "" {
		activeFilamentLabels := []string{"name", "material", "color"}
		activeFilamentDesc := prometheus.NewDesc("klipper_mmu_active_filament_info", "Active filament information (always 1)", activeFilamentLabels, nil)
		ch <- prometheus.MustNewConstMetric(
			activeFilamentDesc,
			prometheus.GaugeValue,
			1,
			mmu.ActiveFilament.FilamentName,
			mmu.ActiveFilament.Material,
			mmu.ActiveFilament.Color)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mmu_active_filament_temperature", "Active filament temperature", nil, nil),
			prometheus.GaugeValue,
			float64(mmu.ActiveFilament.Temperature))

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("klipper_mmu_active_filament_spool_id", "Active filament Spoolman spool ID", nil, nil),
			prometheus.GaugeValue,
			float64(mmu.ActiveFilament.SpoolId))
	}
}

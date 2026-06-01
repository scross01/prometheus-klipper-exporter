package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#history-apis

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type MoonrakerHistoryResponse struct {
	Result struct {
		JobTotals struct {
			Jobs         float64 `json:"total_jobs"`
			TotalTime    float64 `json:"total_time"`
			PrintTime    float64 `json:"total_print_time"`
			FilamentUsed float64 `json:"total_filament_used"`
			LongestJob   float64 `json:"longest_job"`
			LongestPrint float64 `json:"longest_print"`
		} `json:"job_totals"`
	} `json:"result"`
}

type MoonrakerHistoryCurrentPrintResponse struct {
	Result struct {
		Count int `json:"count"`
		Jobs  []struct {
			EndTime      float64 `json:"end_time"`
			FilamentUsed float64 `json:"filament_used"`
			Filename     string  `json:"filename"`
			Metadata     struct {
				Size             int     `json:"size"`
				Modified         float64 `json:"modified"`
				Slicer           string  `json:"slicer"`
				SlicerVersion    string  `json:"slicer_version"`
				LayerHeight      float64 `json:"layer_height"`
				FirstLayerHeight float64 `json:"first_layer_height"`
				ObjectHeight     float64 `json:"object_height"`
				FilamentTotal    float64 `json:"filament_total"`
				EstimatedTime    float64 `json:"estimated_time"`
				Thumbnails       []struct {
					Width        int    `json:"width"`
					Height       int    `json:"height"`
					Size         int    `json:"size"`
					RelativePath string `json:"relative_path"`
				} `json:"thumbnails"`
				FirstLayerBedTemp  float64 `json:"first_layer_bed_temp"`
				FirstLayerExtrTemp float64 `json:"first_layer_extr_temp"`
				GcodeStartByte     int     `json:"gcode_start_byte"`
				GcodeEndByte       int     `json:"gcode_end_byte"`
			} `json:"metadata"`
			PrintDuration float64 `json:"print_duration"`
			Status        string  `json:"status"`
			StartTime     float64 `json:"start_time"`
			TotalDuration float64 `json:"total_duration"`
			JobID         string  `json:"job_id"`
			Exists        bool    `json:"exists"`
		} `json:"jobs"`
	} `json:"result"`
}

func (c Collector) collectActivePrint(ch chan<- prometheus.Metric) {
	var result MoonrakerHistoryCurrentPrintResponse
	if err := c.fetchFromMoonraker("/server/history/list?limit=1&start=0&since=1&order=desc", &result); err != nil {
		log.Error(err)
		return
	}

	if len(result.Result.Jobs) < 1 {
		log.Info("No active print in Current Print repsonse, skipping current print metrics")
	} else {
		c.emitGauge(ch, "klipper_current_print_object_height", "Klipper current print object height", c.checkConditionStatusPrint(result, result.Result.Jobs[0].Metadata.ObjectHeight))
		c.emitGauge(ch, "klipper_current_print_first_layer_height", "Klipper current print first layer height", c.checkConditionStatusPrint(result, result.Result.Jobs[0].Metadata.FirstLayerHeight))
		c.emitGauge(ch, "klipper_current_print_layer_height", "Klipper current print layer height", c.checkConditionStatusPrint(result, result.Result.Jobs[0].Metadata.LayerHeight))
		c.emitGauge(ch, "klipper_current_print_total_duration", "Klipper current print total duration", c.checkConditionStatusPrint(result, result.Result.Jobs[0].TotalDuration))
	}
}

func (c Collector) collectHistory(ch chan<- prometheus.Metric) {
	var result MoonrakerHistoryResponse
	if err := c.fetchFromMoonraker("/server/history/totals", &result); err != nil {
		log.Error(err)
		return
	}
	c.emitGauge(ch, "klipper_total_jobs", "Klipper number of total jobs.", float64(result.Result.JobTotals.Jobs))
	c.emitGauge(ch, "klipper_total_time", "Klipper total time.", result.Result.JobTotals.TotalTime)
	c.emitGauge(ch, "klipper_total_print_time", "Klipper total print time.", result.Result.JobTotals.PrintTime)
	c.emitGauge(ch, "klipper_total_filament_used", "Klipper total meters of filament used.", result.Result.JobTotals.FilamentUsed)
	c.emitGauge(ch, "klipper_longest_job", "Klipper total longest job.", result.Result.JobTotals.LongestJob)
	c.emitGauge(ch, "klipper_longest_print", "Klipper total longest print.", result.Result.JobTotals.LongestPrint)
}

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
	log.Infof("Collecting active print for %s", c.target)

	var result MoonrakerHistoryCurrentPrintResponse
	if err := c.fetchFromMoonraker("/server/history/list?limit=1&start=0&since=1&order=desc", &result); err != nil {
		log.Error(err)
		return
	}

	if len(result.Result.Jobs) < 1 {
		log.Info("No active print in Current Print repsonse, skipping current print metrics")
	} else {
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

func (c Collector) collectHistory(ch chan<- prometheus.Metric) {
	log.Infof("Collecting history for %s", c.target)

	var result MoonrakerHistoryResponse
	if err := c.fetchFromMoonraker("/server/history/totals", &result); err != nil {
		log.Error(err)
		return
	}
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

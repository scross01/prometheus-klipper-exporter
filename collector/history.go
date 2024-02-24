package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#history-apis

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type MoonrakerHistoryResponse struct {
	Result struct {
		JobTotals struct {
			Jobs         int64   `json:"total_jobs"`
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

func (c Collector) fetchMoonrakerHistory(klipperHost string, apiKey string) (*MoonrakerHistoryResponse, error) {
	var url = "http://" + klipperHost + "/server/history/totals"
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

	var response MoonrakerHistoryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &response, nil
}
func (c Collector) fetchMoonrakerHistoryCurrent(klipperHost string, apiKey string) (*MoonrakerHistoryCurrentPrintResponse, error) {
	var url = "http://" + klipperHost + "/server/history/list?limit=1&start=0&since=1&order=desc"
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

	var response MoonrakerHistoryCurrentPrintResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &response, nil
}

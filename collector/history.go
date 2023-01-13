package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#history-apis

import (
	"encoding/json"
	"io/ioutil"

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

func (c collector) fetchMoonrakerHistory(klipperHost string, apiKey string) (*MoonrakerHistoryResponse, error) {
	var url = "http://" + klipperHost + "/server/history/totals"
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
		c.logger.Error(err)
		return nil, err
	}

	var response MoonrakerHistoryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	return &response, nil
}

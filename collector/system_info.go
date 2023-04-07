package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#get-system-info

import (
	"encoding/json"
	"io/ioutil"

	"net/http"
)

type MoonrakerSystemInfoQueryResponse struct {
	Result struct {
		SystemInfo struct {
			CpuInfo struct {
				CpuCount    int    `json:"cpu_count"`
				TotalMemory int    `json:"total_memory"`
				MemoryUnits string `json:"memory_units"`
			} `json:"cpu_info"`
		} `json:"system_info"`
	} `json:"result"`
}

func (c collector) fetchMoonrakerSystemInfo(klipperHost string, apiKey string) (*MoonrakerSystemInfoQueryResponse, error) {
	var url = "http://" + klipperHost + "/machine/system_info"
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

	var response MoonrakerSystemInfoQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}

	return &response, nil
}

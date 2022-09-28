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
			CpuInfo	struct { 
				CpuCount    int    `json:"cpu_count"`
				TotalMemory int    `json:"total_memory"`
				MemoryUnits string `json:"memory_units"`	
			} `json:"cpu_info"`	
		} `json:"system_info"`
	} `json:"result"`
}

func (c collector)fetchMoonrakerSystemInfo(klipperHost string) (*MoonrakerSystemInfoQueryResponse, error) {
	var procStatsUrl = "http://" + klipperHost + "/machine/system_info"
	c.logger.Info("Collecting metrics from " + procStatsUrl)
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

	var response MoonrakerSystemInfoQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		c.logger.Fatal(err)
		return nil, err
	}

	return &response, nil
}

package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#get-directory-information

import (
	"encoding/json"
	"io/ioutil"

	"net/http"
)

type MoonrakerDirecotryInfoQueryResponse struct {
	Result struct {
		DiskUsage struct { 
			Total int64 `json:"total"`
			Used  int64 `json:"used"`
			Free  int64 `json:"free"`	
		} `json:"disk_usage"`	
	} `json:"result"`
}

func (c collector)fetchMoonrakerDirectoryInfo(klipperHost string) (*MoonrakerDirecotryInfoQueryResponse, error) {
	var procStatsUrl = "http://" + klipperHost + "/server/files/directory?path=gcodes&extended=false"
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

	var response MoonrakerDirecotryInfoQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		c.logger.Fatal(err)
		return nil, err
	}

	return &response, nil
}

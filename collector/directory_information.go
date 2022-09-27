package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#get-directory-information

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
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

func fetchMoonrakerDirectoryInfo(klipperHost string) (*MoonrakerDirecotryInfoQueryResponse, error) {
	var procStatsUrl = "http://" + klipperHost + "/server/files/directory?path=gcodes&extended=false"
	log.Info("Collecting metrics from " + procStatsUrl)
	res, err := http.Get(procStatsUrl)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var response MoonrakerDirecotryInfoQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	log.Info("Collected metrics from " + procStatsUrl)

	return &response, nil
}

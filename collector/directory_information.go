package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#get-directory-information

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
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

func (c Collector) fetchMoonrakerDirectoryInfo(klipperHost string, apiKey string) (*MoonrakerDirecotryInfoQueryResponse, error) {
	var url = "http://" + klipperHost + "/server/files/directory?path=gcodes&extended=false"
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

	var response MoonrakerDirecotryInfoQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &response, nil
}

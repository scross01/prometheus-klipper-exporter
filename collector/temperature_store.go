package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#request-cached-temperature-data

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type TemperatureDataQueryResponse struct {
	Result map[string]interface{} `json:"result"`
}

func (c Collector) fetchTemperatureData(klipperHost string, apiKey string) (*TemperatureDataQueryResponse, error) {
	var url = "http://" + klipperHost + "/server/temperature_store"
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

	var response TemperatureDataQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &response, nil
}

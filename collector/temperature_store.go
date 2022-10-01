package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#request-cached-temperature-data

import (
	"encoding/json"
	"io/ioutil"

	"net/http"
)

type TemperatureDataQueryResponse struct {
	Result map[string]interface{} `json:"result"`
}

func (c collector) fetchTemperatureData(klipperHost string) (*TemperatureDataQueryResponse, error) {
	var procStatsUrl = "http://" + klipperHost + "/server/temperature_store"
	c.logger.Debug("Collecting metrics from " + procStatsUrl)
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

	var response TemperatureDataQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		c.logger.Fatal(err)
		return nil, err
	}

	return &response, nil
}

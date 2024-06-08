package collector

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type SpoolmanStatusQueryResponse struct {
	Result struct {
		SpoolmanConnected bool `json:"spoolman_connected"`
		PendingReports    []struct {
			SpoolId      int     `json:"spool_id"`
			FilamentUsed float64 `json:"filament_used"`
		} `json:"pending_reports"`
		SpoolId int `json:"spool_id"`
	} `json:"result"`
}

func (c Collector) fetchSpoolmanStatus(klipperHost string, apiKey string) (*SpoolmanStatusQueryResponse, error) {
	var url = "http://" + klipperHost + "/server/spoolman/status"
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

	var response SpoolmanStatusQueryResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &response, nil
}

package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#retrieve-the-job-queue-status

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	log "github.com/sirupsen/logrus"
)

type MoonrakerJobQueueResponse struct {
	Result struct {
		QueuedJobs []MoonrakerQueuedJob `json:"queued_jobs"`
		QueueState string               `json:"queue_state"`
	} `json:"result"`
}

type MoonrakerQueuedJob struct {
	TimeInQueue float64 `json:"time_in_queue"`
}

func (c Collector) fetchMoonrakerJobQueue(klipperHost string, apiKey string) (*MoonrakerJobQueueResponse, error) {
	var url = "http://" + klipperHost + "/server/job_queue/status"
	log.Debug("Collecting metrics from " + url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP request for %s. %s", url, err)
	}
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to complete HTTP client request. %s", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read data from HTTP response. %s", err)
	}

	var response MoonrakerJobQueueResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response data to %s. %s", reflect.TypeOf(response), err)
	}

	return &response, nil
}

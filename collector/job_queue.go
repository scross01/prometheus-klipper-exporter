package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#retrieve-the-job-queue-status

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"net/http"
)

type MoonrakerJobQueueResponse struct {
	Result struct {
		QueuedJobs  []MoonrakerQueuedJob `json:"queued_jobs"`
		QueueState 	string               `json:"queue_state"`
	} `json:"result"`
}

type MoonrakerQueuedJob struct {
	TimeInQueue int `json:"time_in_queue"`	
}

func fetchMoonrakerJobQueue(klipperHost string) (*MoonrakerJobQueueResponse, error) {
	var procStatsUrl = "http://" + klipperHost + "/server/job_queue/status"
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

	var response MoonrakerJobQueueResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	log.Info("Collected metrics from " + procStatsUrl)

	return &response, nil
}

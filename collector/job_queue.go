package collector

// https://moonraker.readthedocs.io/en/latest/web_api/#retrieve-the-job-queue-status

import (
	"github.com/prometheus/client_golang/prometheus"
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

func (c Collector) collectJobQueue(ch chan<- prometheus.Metric) {
	var result MoonrakerJobQueueResponse
	if err := c.fetchFromMoonraker("/server/job_queue/status", &result); err != nil {
		log.Error(err)
		return
	}

	c.emitGauge(ch, "klipper_job_queue_length", "Klipper job queue length.", float64(len(result.Result.QueuedJobs)))
	emitStateInfoMetric(ch, "klipper_job_queue_state_info", "The current state of the job queue.", "state", result.Result.QueueState)
}

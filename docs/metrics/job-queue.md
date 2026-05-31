# Job Queue

**Module:** `job_queue` (default)  
**API Endpoint:** [`/server/job_queue/status`](https://moonraker.readthedocs.io/en/latest/web_api/#retrieve-the-job-queue-status)

Collects the current job queue length from Moonraker.

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_job_queue_length` | Gauge | Number of jobs currently in the queue |

## Example PromQL

```promql
# Current queue length
klipper_job_queue_length

# Alert if queue is growing
rate(klipper_job_queue_length[5m]) > 0
```

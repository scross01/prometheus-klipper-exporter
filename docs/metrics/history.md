# History

**Module:** `history` (optional)  
**API Endpoints:**
- [`/server/history/totals`](https://moonraker.readthedocs.io/en/latest/web_api/#get-history-totals) — aggregate job statistics
- [`/server/history/list`](https://moonraker.readthedocs.io/en/latest/web_api/#get-history-list) — current/active print metadata

Collects historical print job statistics and current in-progress print metadata.

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_total_jobs` | Gauge | Total number of print jobs completed |
| `klipper_total_time` | Gauge | Total elapsed time across all jobs in seconds |
| `klipper_total_print_time` | Gauge | Total printing time across all jobs in seconds |
| `klipper_total_filament_used` | Gauge | Total filament used across all jobs |
| `klipper_longest_job` | Gauge | Duration of the longest job in seconds |
| `klipper_longest_print` | Gauge | Duration of the longest print in seconds |
| `klipper_current_print_first_layer_height` | Gauge | First layer height of the current in-progress print |
| `klipper_current_print_layer_height` | Gauge | Layer height of the current in-progress print |
| `klipper_current_print_object_height` | Gauge | Object height of the current in-progress print |
| `klipper_current_print_total_duration` | Gauge | Total duration of the current in-progress print in seconds |

> Current print metrics (`klipper_current_print_*`) only report values when a
> print is in progress (status is `in_progress`). When idle, they report 0.

## Example PromQL

```promql
# Total jobs completed
klipper_total_jobs

# Average print time
klipper_total_print_time / klipper_total_jobs

# Current print progress (if using slicer estimated time)
klipper_current_print_total_duration
```

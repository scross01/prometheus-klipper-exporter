# Directory Info

**Module:** `directory_info` (optional)  
**API Endpoint:** [`/server/files/directory`](https://moonraker.readthedocs.io/en/latest/web_api/#get-directory-information)

Collects disk usage statistics for the gcodes directory on the Klipper host.

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_disk_usage_available` | Gauge | Available disk space in bytes |
| `klipper_disk_usage_total` | Gauge | Total disk space in bytes |
| `klipper_disk_usage_used` | Gauge | Used disk space in bytes |

## Example PromQL

```promql
# Disk usage percentage
(klipper_disk_usage_used / klipper_disk_usage_total) * 100

# Available space in GB
klipper_disk_usage_available / 1073741824
```

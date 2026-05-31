# System Info

**Module:** `system_info` (default)  
**API Endpoint:** [`/machine/system_info`](https://moonraker.readthedocs.io/en/latest/web_api/#get-system-info)

Collects system information from the Klipper host.

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_system_cpu_count` | Gauge | Number of CPU cores |

## Example PromQL

```promql
# CPU count
klipper_system_cpu_count
```

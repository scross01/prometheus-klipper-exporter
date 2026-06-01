# System Info

**Module:** `system_info` (default)
**API Endpoint:** [`/machine/system_info`](https://moonraker.readthedocs.io/en/latest/web_api/#get-system-info)

Collects system information from the Klipper host, including CPU details and
Moonraker-managed service states.

## Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_system_cpu_count` | Gauge | | Number of CPU cores |
| `klipper_service_available` | Gauge=1 | `service` | Service is registered as available (always 1 when present) |
| `klipper_service_state_info` | Gauge=1 | `service`, `state` | Current state of the service (e.g. active, inactive, failed) |
| `klipper_service_sub_state_info` | Gauge=1 | `service`, `sub_state` | Current sub-state of the service (e.g. running, dead, exited) |

## Example PromQL

```promql
# CPU count
klipper_system_cpu_count

# Count available services
count(klipper_service_available)

# Services that are in a failed state
klipper_service_state_info{state="failed"}

# Check if klipper service is running
klipper_service_sub_state_info{service="klipper", sub_state="running"}
```

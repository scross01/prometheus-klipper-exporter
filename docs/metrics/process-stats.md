# Process Stats

**Module:** `process_stats` (default)  
**API Endpoint:** [`/machine/proc_stats`](https://moonraker.readthedocs.io/en/latest/web_api/#get-moonraker-process-stats)

Collects Moonraker process statistics and system-level CPU, memory, and uptime
metrics from the Klipper host.

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_moonraker_cpu_usage` | Gauge | Moonraker process CPU usage |
| `klipper_moonraker_memory_kb` | Gauge | Moonraker process memory usage in KB |
| `klipper_moonraker_websocket_connections` | Gauge | Number of active Moonraker websocket connections |
| `klipper_system_cpu` | Gauge | System-wide CPU usage percentage |
| `klipper_system_cpu_temp` | Gauge | System CPU temperature in Celsius |
| `klipper_system_memory_available` | Gauge | Available system memory |
| `klipper_system_memory_total` | Gauge | Total system memory |
| `klipper_system_memory_used` | Gauge | Used system memory |
| `klipper_system_uptime` | Counter | System uptime in seconds |

## Example PromQL

```promql
# Moonraker memory usage over time
klipper_moonraker_memory_kb

# System CPU temperature
klipper_system_cpu_temp

# System uptime in days
klipper_system_uptime / 86400
```

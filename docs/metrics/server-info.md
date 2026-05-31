# Server Info

**Module:** `server_info` (default)  
**API Endpoint:** [`/server/info`](https://moonraker.readthedocs.io/en/latest/web_api/#get-moonraker-server-info)

Collects Moonraker server information including Klippy connection state, loaded
components, and version details.

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_klippy_connected` | Gauge | Whether Klippy is connected (1) or not (0) |
| `klipper_klippy_state_info` | Gauge=1 | The current state of Klippy with `state` label (ready, error, shutdown, startup, disconnected) |
| `klipper_component_info` | Gauge=1 | A registered Moonraker component with `component` label |
| `klipper_component_failed_info` | Gauge=1 | A Moonraker component that failed to load with `failed_component` label |
| `klipper_moonraker_version_info` | Gauge=1 | Moonraker version with `version` label |
| `klipper_api_version_info` | Gauge=1 | Moonraker API version with `version` label |

## Example PromQL

```promql
# Is Klippy connected?
klipper_klippy_connected

# Current Klippy state
klipper_klippy_state_info

# Failed components
klipper_component_failed_info
```

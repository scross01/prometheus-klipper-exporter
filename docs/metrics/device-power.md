# Device Power

**Module:** `device_power` (default)
**API Endpoints:** [`/machine/device_power/devices`](https://moonraker.readthedocs.io/en/latest/external_api/devices/) and [`/machine/device_power/status`](https://moonraker.readthedocs.io/en/latest/external_api/devices/)

Monitors the state of configurable power devices (smart plugs, GPIO relays, etc.)
defined in `moonraker.conf` under `[power <device_name>]` sections.

## Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_power_device_info` | Gauge=1 | `device`, `type` | Power device information with hardware type |
| `klipper_power_device_status` | Gauge | `device` | Device on/off status (1=on, 0=off/error/init) |
| `klipper_power_device_state_info` | Gauge=1 | `device`, `state` | Current device state (`on`, `off`, `error`, `init`) |

## Example PromQL

```promql
# Check if a specific device is on
klipper_power_device_status{device="printer"}

# Show current state of all power devices
klipper_power_device_state_info

# Count of devices that are currently on
count(klipper_power_device_status == 1)
```

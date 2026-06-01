# Spoolman

**Module:** `spoolman` (optional)
**API Endpoint:** [`POST /server/spoolman/proxy` → `GET /v1/spool`](https://moonraker.readthedocs.io/en/latest/external_api/integrations/#proxy)Spoolman is a filament spool manager that Moonraker can proxy. This module
exposes metrics for each spool tracked in Spoolman, including remaining and
used filament weight and length, as well as Spoolman connection status and
active spool tracking.

Spool data is fetched through Moonraker's [Spoolman proxy](https://moonraker.readthedocs.io/en/latest/external_api/integrations/#proxy)
using the v2 response format with `use_v2_response: true`. Connection status
and active spool info are fetched from the [`/server/spoolman/status`](https://moonraker.readthedocs.io/en/latest/external_api/integrations/#get-spoolman-status)
endpoint.

### Status Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_spoolman_connected` | Gauge | Spoolman connection status (1=connected, 0=disconnected) |
| `klipper_spoolman_active_spool_id` | Gauge | Currently active spool ID (-1 if no spool is active) |
| `klipper_spoolman_pending_reports` | Gauge | Number of pending filament usage reports not yet sent to Spoolman |

### Per-Spool Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_spoolman_spool_info` | Gauge=1 | `spool_id`, `filament_name`, `material`, `color`, `vendor` | Spool information (always 1) |
| `klipper_spoolman_remaining_weight` | Gauge | `spool_id` | Remaining filament weight on the spool (grams) |
| `klipper_spoolman_used_weight` | Gauge | `spool_id` | Used filament weight from the spool (grams) |
| `klipper_spoolman_remaining_length` | Gauge | `spool_id` | Remaining filament length on the spool (mm) |
| `klipper_spoolman_used_length` | Gauge | `spool_id` | Used filament length from the spool (mm) |

## Example PromQL

```promql
# Show all spools with their filament names
klipper_spoolman_spool_info

# Remaining weight for a specific spool
klipper_spoolman_remaining_weight{spool_id="1"}

# Spools with less than 100g remaining
klipper_spoolman_remaining_weight < 100
```

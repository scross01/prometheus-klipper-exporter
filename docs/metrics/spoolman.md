# Spoolman

**Module:** `spoolman` (optional)
**API Endpoint:** [`/server/spoolman/spool`](https://moonraker.readthedocs.io/en/latest/external_api/integrations/#spoolman)

Spoolman is a filament spool manager that Moonraker can proxy. This module
exposes metrics for each spool tracked in Spoolman, including remaining and
used filament weight and length.

## Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_spoolman_spool_info` | Gauge=1 | `spool_id`, `filament_name`, `material`, `color` | Spool information (always 1) |
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

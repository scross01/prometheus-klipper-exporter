# CFS — Creality Filament System

**Module:** `cfs` (optional)
**API Endpoint:** [`/printer/objects/query`](https://moonraker.readthedocs.io/en/latest/web_api/#query-printer-object-status)
with `box`, `filament_rack`, and `load_ai` objects.

Collects metrics from the **Creality Filament System** (CFS) on K2-class printers.
These printers do **not** populate Happy Hare's `mmu` object, so the [`mmu`](./mmu)
module reports nothing on them. The CFS state instead lives in native Moonraker
objects:

- `box` — the CFS unit(s) and per-slot state (active slot, temperature, humidity)
- `filament_rack` — the filament currently loaded at the toolhead
- `load_ai` — Creality's AI print-defect detection

Up to four units (`T1`–`T4`) are reported. Disconnected units (those with
`state == "None"`) are skipped, so only attached units emit `unit`-labelled series.

## Metrics

### Box State

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_cfs_enabled` | Gauge | | CFS enabled state (0/1) |
| `klipper_cfs_auto_refill_enabled` | Gauge | | Auto-refill enabled (0/1) |
| `klipper_cfs_filament_useup` | Gauge | | Filament used-up flag (0/1) |
| `klipper_cfs_state_info` | Gauge=1 | `state` | CFS connection state |
| `klipper_cfs_active_unit` | Gauge | | `box.filament` value (semantics unconfirmed; likely active unit number) |

### Active Slot

Labels: `unit`

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_cfs_active_slot` | Gauge | `unit` | Active slot index within the unit (A=0..D=3, -1 if none) |
| `klipper_cfs_active_slot_info` | Gauge=1 | `unit`, `slot`, `material`, `color` | Active slot details (emitted only when a slot is active) |

### Per-Unit Metrics

Labels: `unit`

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_cfs_unit_state_info` | Gauge=1 | `unit`, `state` | Unit connection state |
| `klipper_cfs_unit_temperature_celsius` | Gauge | `unit` | Unit temperature in °C |
| `klipper_cfs_unit_humidity_percent` | Gauge | `unit` | Unit relative humidity (assumed %RH) |
| `klipper_cfs_unit_info` | Gauge=1 | `unit`, `version`, `sn`, `mode` | Unit hardware info |

### Per-Slot Metrics

Labels: `unit`, `slot` (slot letter A–D)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_cfs_slot_info` | Gauge=1 | `unit`, `slot`, `material`, `color`, `vendor` | Slot details (`material` is the raw type code) |
| `klipper_cfs_slot_remaining` | Gauge | `unit`, `slot` | Remaining filament (units unclear: percent or mm) |

### Filament Rack (loaded at toolhead)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_cfs_rack_loaded_info` | Gauge=1 | `material`, `color` | Filament currently loaded at the toolhead |
| `klipper_cfs_rack_velocity` | Gauge | | Loaded filament velocity (units unclear, likely mm/min) |

### AI Print-Defect Detection

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_cfs_ai_detection_enabled` | Gauge | | AI print-defect detection enabled (0/1) |
| `klipper_cfs_ai_waste_detection_enabled` | Gauge | | AI waste detection enabled (0/1) |
| `klipper_cfs_ai_max_probability` | Gauge | | Maximum defect recognition probability |
| `klipper_cfs_ai_normalized_area` | Gauge | | Normalized total defect area |
| `klipper_cfs_ai_command_info` | Gauge=1 | `command_type` | Current AI command type (emitted only when non-empty) |

## Notes & Caveats

Some fields are emitted with conservative names because their exact semantics are
not yet confirmed against the printer UI:

- `klipper_cfs_active_unit` (`box.filament`) — could be an active unit number, a
  count of loaded filaments, or a boolean. Indistinguishable with a single unit
  attached.
- `klipper_cfs_slot_remaining` (`remain_len`) — units unknown (percent vs mm);
  reads `100` on a freshly loaded spool.
- `klipper_cfs_unit_humidity_percent` (`dry_and_humidity`) — assumed %RH.
- `material` labels carry the raw Creality material **code** (e.g. `000001`), not
  the human-readable name (e.g. `PLA`).

## Example PromQL

```promql
# Active slot index per CFS unit (A=0..D=3)
klipper_cfs_active_slot

# Material/color of the active slot
klipper_cfs_active_slot_info

# Unit humidity (consider drying when high)
klipper_cfs_unit_humidity_percent > 40

# AI defect detection confidence
klipper_cfs_ai_max_probability
```

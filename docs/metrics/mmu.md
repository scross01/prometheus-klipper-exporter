# MMU — Multi-Material Unit

**Module:** `mmu` (optional)  
**API Endpoint:** [`/printer/objects/query`](https://moonraker.readthedocs.io/en/latest/web_api/#query-printer-object-status)
with `mmu`, `mmu_encoder mmu_encoder`, and `mmu_machine` objects.

Collects metrics from [Happy Hare](https://github.com/moggieuk/Happy-Hare)
Multi-Material Unit — an advanced filament management system for Klipper.

## Metrics

### Basic State

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_mmu_enabled` | Gauge | MMU enabled state (0/1) |
| `klipper_mmu_homed` | Gauge | MMU homed state (0/1) |
| `klipper_mmu_num_gates` | Gauge | Number of MMU gates |
| `klipper_mmu_has_bypass` | Gauge | Whether MMU has a bypass gate (0/1) |
| `klipper_mmu_current_unit` | Gauge | Current MMU unit index |
| `klipper_mmu_current_tool` | Gauge | Current tool (-1=unknown, -2=bypass) |
| `klipper_mmu_current_gate` | Gauge | Current gate index |

### Print State & Action

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_mmu_print_state_info` | Gauge=1 | `state` | MMU print state |
| `klipper_mmu_action_info` | Gauge=1 | `action` | Current MMU action |
| `klipper_mmu_operation_info` | Gauge=1 | `operation` | Current MMU operation |
| `klipper_mmu_sync_feedback_state_info` | Gauge=1 | `state` | Sync feedback state |

### Filament

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_mmu_filament_loaded` | Gauge | Whether filament is loaded (0/1) |
| `klipper_mmu_filament_position_mm` | Gauge | Filament position in mm |
| `klipper_mmu_filament_pos_state` | Gauge | Filament position state machine value |
| `klipper_mmu_filament_direction` | Gauge | Filament direction (1=load, -1=unload) |
| `klipper_mmu_runout` | Gauge | Runout detected (0/1) |

### Toolchange

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_mmu_toolchanges_total` | Gauge | Total toolchanges in current print |
| `klipper_mmu_last_tool` | Gauge | Last tool used |
| `klipper_mmu_next_tool` | Gauge | Next tool during toolchange |
| `klipper_mmu_toolchange_purge_volume_mm3` | Gauge | Suggested purge volume in mm³ |
| `klipper_mmu_slicer_total_toolchanges` | Gauge | Expected toolchanges from slicer |
| `klipper_mmu_slicer_initial_tool` | Gauge | Initial tool from slicer |

### Detection & Sync

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_mmu_clog_detection_mode` | Gauge | Clog detection mode (0=off, 1=manual, 2=auto) |
| `klipper_mmu_endless_spool_enabled` | Gauge | Endless spool mode (0=off, 1=enabled, 2=pre-gate) |
| `klipper_mmu_sync_drive_enabled` | Gauge | Gear stepper synced to extruder (0/1) |
| `klipper_mmu_servo_position_info` | Gauge=1 | Servo position as labeled gauge |
| `klipper_mmu_bowden_progress_percent` | Gauge | Bowden move progress (-1 if inactive) |

### Encoder

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_mmu_encoder_position_mm` | Gauge | Encoder position in mm |
| `klipper_mmu_encoder_detection_length_mm` | Gauge | Clog detection length in mm |
| `klipper_mmu_encoder_headroom_mm` | Gauge | Current clog detection headroom |
| `klipper_mmu_encoder_min_headroom_mm` | Gauge | Minimum recorded headroom |
| `klipper_mmu_encoder_desired_headroom_mm` | Gauge | Desired headroom |
| `klipper_mmu_encoder_flow_rate_percent` | Gauge | Encoder flow rate percentage |
| `klipper_mmu_encoder_enabled` | Gauge | Encoder enabled for clog detection (0/1) |

### Per-Gate Metrics

Labels: `gate`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_mmu_gate_status` | Gauge | Status (-1=unknown, 0=empty, 1=available, 2=buffered) |
| `klipper_mmu_gate_temperature` | Gauge | Gate filament temperature |
| `klipper_mmu_gate_speed_override_percent` | Gauge | Speed override percentage |
| `klipper_mmu_gate_ttg_map` | Gauge | Tool-to-gate mapping |
| `klipper_mmu_gate_endless_spool_group` | Gauge | Endless spool group |
| `klipper_mmu_gate_spool_id` | Gauge | Spoolman spool ID (-1 if unset) |
| `klipper_mmu_gate_info` | Gauge=1 | Gate info with `gate`, `material`, `color`, `filament_name` labels |
| `klipper_mmu_pre_gate_sensor_detected` | Gauge | Pre-gate filament detected (0/1) |
| `klipper_mmu_pre_gate_sensor_enabled` | Gauge | Pre-gate sensor enabled (0/1) |

### Tool Multipliers

Labels: `tool`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_mmu_tool_extrusion_multiplier` | Gauge | Tool extrusion multiplier (M221) |
| `klipper_mmu_tool_speed_multiplier` | Gauge | Tool speed multiplier (M220) |

### Machine & Active Filament Info

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_mmu_machine_info` | Gauge=1 | `name`, `vendor`, `version`, `selector_type` | MMU machine info |
| `klipper_mmu_num_units` | Gauge | | Number of MMU units |
| `klipper_mmu_active_filament_info` | Gauge=1 | `name`, `material`, `color` | Active filament info |
| `klipper_mmu_active_filament_temperature` | Gauge | | Active filament temperature |
| `klipper_mmu_active_filament_spool_id` | Gauge | | Active filament Spoolman spool ID |

## Example PromQL

```promql
# Current gate and tool
klipper_mmu_current_gate
klipper_mmu_current_tool

# Gate status (1 = available)
klipper_mmu_gate_status == 1

# Toolchanges per print
klipper_mmu_toolchanges_total

# Encoder headroom (clog risk when low)
klipper_mmu_encoder_headroom_mm < 5
```

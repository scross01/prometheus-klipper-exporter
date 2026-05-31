# Printer Objects

**Module:** `printer_objects` (optional)  
**API Endpoint:** [`/printer/objects/query`](https://moonraker.readthedocs.io/en/latest/web_api/#query-printer-object-status)

Collects state from Klipper printer objects. This is the largest and most
important module, covering temperatures, fans, motors, toolhead, print status,
and more. Objects with dynamic instances (e.g., `temperature_sensor my_sensor`)
are auto-discovered via [`/printer/objects/list`](https://moonraker.readthedocs.io/en/latest/web_api/#retrieve-a-list-of-printers-available-printer-objects).

---

### `webhooks`

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_webhooks_state_info` | Gauge=1 | `state` | Webhooks server state (`ready`, `startup`, `shutdown`, `error`) |

---

### `pause_resume`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_pause_resume_is_paused` | Gauge | Whether the print is paused (1) or not (0) |

---

### `controller_fan`

Labels: `fan`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_controller_fan_rpm` | Gauge | Controller fan RPM |
| `klipper_controller_fan_speed` | Gauge | Controller fan speed (0–1) |

---

### `display_status`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_print_gcode_progress` | Gauge | Print progress percentage as reported by M73 |

---

### `extruder`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_extruder_power` | Gauge | Extruder heater power (0–1) |
| `klipper_extruder_pressure_advance` | Gauge | Pressure advance value |
| `klipper_extruder_smooth_time` | Gauge | Pressure advance smooth time |
| `klipper_extruder_target` | Gauge | Target extruder temperature |
| `klipper_extruder_temperature` | Gauge | Current extruder temperature |

---

### `fan`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_fan_rpm` | Gauge | Part cooling fan RPM |
| `klipper_fan_speed` | Gauge | Part cooling fan speed (0–1) |

---

### `filament_sensor`

Labels: `sensor`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_filament_sensor_detected` | Gauge | Whether filament is detected (0/1) |
| `klipper_filament_sensor_enabled` | Gauge | Whether the sensor is enabled (0/1) |

---

### `gcode`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_gcode_extrude_factor` | Gauge | Extrude multiplier (M221) |
| `klipper_gcode_position_e` | Gauge | Current E position in G-code coordinates |
| `klipper_gcode_position_x` | Gauge | Current X position in G-code coordinates |
| `klipper_gcode_position_y` | Gauge | Current Y position in G-code coordinates |
| `klipper_gcode_position_z` | Gauge | Current Z position in G-code coordinates |
| `klipper_gcode_speed` | Gauge | Current G-code speed |
| `klipper_gcode_speed_factor` | Gauge | Speed multiplier override (M220) |

---

### `generic_fan`

Labels: `fan`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_generic_fan_rpm` | Gauge | Generic fan RPM |
| `klipper_generic_fan_speed` | Gauge | Generic fan speed (0–1) |

---

### `heater_bed`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_heater_bed_power` | Gauge | Heated bed power (0–1) |
| `klipper_heater_bed_target` | Gauge | Target bed temperature |
| `klipper_heater_bed_temperature` | Gauge | Current bed temperature |

---

### `heater_fan`

Labels: `fan`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_heater_fan_rpm` | Gauge | Heater fan RPM |
| `klipper_heater_fan_speed` | Gauge | Heater fan speed (0–1) |

---

### `heater_generic`

Labels: `heater`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_generic_heater_power` | Gauge | Generic heater power (0–1) |
| `klipper_generic_heater_target` | Gauge | Target temperature |
| `klipper_generic_heater_temperature` | Gauge | Current temperature |

---

### `idle_timeout`

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_printing_time` | Counter | — | Time spent in the Printing state |
| `klipper_idle_timeout_state_info` | Gauge=1 | `state` | Idle timeout state (`Idle`, `Printing`, `Ready`) |

---

### `mcu`

Labels: `mcu`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_mcu_awake` | Gauge | MCU awake status |
| `klipper_mcu_task_avg` | Gauge | MCU average task time |
| `klipper_mcu_task_stddev` | Gauge | MCU task time standard deviation |
| `klipper_mcu_clock_frequency` | Gauge | MCU clock frequency |
| `klipper_mcu_invalid_bytes` | Gauge | Invalid bytes received |
| `klipper_mcu_read_bytes` | Gauge | Bytes read from MCU |
| `klipper_mcu_ready_bytes` | Gauge | Ready buffer size |
| `klipper_mcu_receive_seq` | Gauge | Receive sequence number |
| `klipper_mcu_retransmit_bytes` | Gauge | Retransmitted bytes |
| `klipper_mcu_retransmit_seq` | Gauge | Retransmit sequence number |
| `klipper_mcu_rto` | Gauge | Retransmission timeout |
| `klipper_mcu_rttvar` | Gauge | Round trip time variance |
| `klipper_mcu_send_seq` | Gauge | Send sequence number |
| `klipper_mcu_stalled_bytes` | Gauge | Stalled buffer bytes |
| `klipper_mcu_srtt` | Gauge | Smoothed round trip time |
| `klipper_mcu_write_bytes` | Gauge | Bytes written to MCU |

---

### `output_pin`

Labels: `pin`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_output_pin_value` | Gauge | Output pin value |

---

### `print_stats`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_print_filament_used` | Gauge | Filament used in current print (mm) |
| `klipper_print_total_duration` | Gauge | Total elapsed time since print started (seconds) |
| `klipper_print_state_info` | Gauge=1 | Current print state (`standby`, `printing`, `paused`, `error`, `complete`) with `state` label |
| `klipper_printing` | Gauge | Whether printer is actively printing (1) or not (0) |

---

### `query_endstops`

Labels: `endstop`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_endstop_triggered` | Gauge | Whether an endstop is triggered (1) or not (0) |

---

### `temperature_fan`

Labels: `fan`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_temperature_fan_rpm` | Gauge | Temperature-controlled fan RPM |
| `klipper_temperature_fan_speed` | Gauge | Temperature-controlled fan speed (0–1) |
| `klipper_temperature_fan_temperature` | Gauge | Temperature fan sensor reading |
| `klipper_temperature_fan_target` | Gauge | Temperature fan target threshold |

---

### `temperature_probe`

Labels: `probe`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_temperature_probe_temperature` | Gauge | Current probe temperature |
| `klipper_temperature_probe_measured_max_temp` | Gauge | Maximum measured probe temperature |
| `klipper_temperature_probe_measured_min_temp` | Gauge | Minimum measured probe temperature |
| `klipper_temperature_probe_estimated_expansion` | Gauge | Estimated probe thermal expansion |

---

### `temperature_sensor`

Labels: `sensor`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_temperature_sensor_temperature` | Gauge | Current sensor temperature |
| `klipper_temperature_sensor_measured_max_temp` | Gauge | Maximum measured temperature |
| `klipper_temperature_sensor_measured_min_temp` | Gauge | Minimum measured temperature |

---

### `tmc_sensor`

Labels: `sensor`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_tmc_sensor_enabled` | Gauge | Whether TMC driver is enabled (0/1) |
| `klipper_tmc_sensor_run_current` | Gauge | TMC driver run current |
| `klipper_tmc_sensor_temperature` | Gauge | TMC driver temperature (if available) |

---

### `toolhead`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_toolhead_estimated_print_time` | Gauge | Estimated total print time (seconds) |
| `klipper_toolhead_homed_axes_info` | Gauge=1 | A homed axis with `axis` label |
| `klipper_toolhead_max_accel` | Gauge | Maximum acceleration (mm/s²) |
| `klipper_toolhead_max_accel_to_decel` | Gauge | Maximum acceleration to deceleration |
| `klipper_toolhead_max_velocity` | Gauge | Maximum velocity (mm/s) |
| `klipper_toolhead_print_time` | Gauge | Current print time (seconds) |
| `klipper_toolhead_square_corner_velocity` | Gauge | Square corner velocity (mm/s) |
| `klipper_toolhead_stalls_total` | Counter | Total number of toolhead stalls |

---

### `virtual_sdcard`

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_print_file_position` | Gauge | Current file position in bytes |
| `klipper_print_file_progress` | Gauge | File read progress as percentage |
| `klipper_sdcard_active` | Gauge | Whether the virtual SD card is actively being read (1) or not (0) |

## Example PromQL

```promql
# Current extruder temperature
klipper_extruder_temperature

# Bed heating to target
klipper_heater_bed_temperature / klipper_heater_bed_target

# Print state (is printing?)
klipper_print_state_info{state="printing"}

# Total print time estimate
klipper_toolhead_estimated_print_time

# MCU communication health
klipper_mcu_retransmit_bytes / klipper_mcu_write_bytes > 0.01
```

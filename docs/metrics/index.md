# Metrics Reference

This page provides a summary of all metrics exported by the Prometheus Klipper
Exporter. Each metric module links to a detailed reference page.

## Modules Overview

| Module | Default | API Endpoint | Metrics |
|--------|---------|--------------|---------|
| `server_info` | ✓ | [`/server/info`](../metrics/server-info) | 6 |
| `process_stats` | ✓ | [`/machine/proc_stats`](../metrics/process-stats) | 11 |
| `network_stats` | | [`/machine/proc_stats`](../metrics/network-stats) | 9 |
| `system_info` | ✓ | [`/machine/system_info`](../metrics/system-info) | 1 |
| `job_queue` | ✓ | [`/server/job_queue/status`](../metrics/job-queue) | 2 |
| `directory_info` | | [`/server/files/directory`](../metrics/directory-info) | 3 |
| `history` | | [`/server/history/totals`](../metrics/history) | 10 |
| `printer_objects` | | [`/printer/objects/query`](../metrics/printer-objects) | 80+ |
| `mmu` | | [`/printer/objects/query`](../metrics/mmu) | 50+ |

## Default Modules

When no modules are specified in the Prometheus configuration, the following
default modules are enabled: `server_info`, `process_stats`, `job_queue`, `system_info`.

## All Metrics by Module

### `process_stats`

| Metric | Type | Labels |
|--------|------|--------|
| `klipper_moonraker_cpu_usage` | Gauge | |
| `klipper_moonraker_memory_kb` | Gauge | |
| `klipper_moonraker_websocket_connections` | Gauge | |
| `klipper_system_cpu` | Gauge | |
| `klipper_system_cpu_temp` | Gauge | |
| `klipper_system_memory_available` | Gauge | |
| `klipper_system_memory_total` | Gauge | |
| `klipper_system_memory_used` | Gauge | |
| `klipper_system_uptime` | Counter | |
| `klipper_system_throttled_bits` | Gauge | |
| `klipper_system_throttled_flag_info` | Gauge=1 | `flag` |

[Full reference →](./process-stats)

### `server_info`

| Metric | Type | Labels |
|--------|------|--------|
| `klipper_klippy_connected` | Gauge | |
| `klipper_klippy_state_info` | Gauge=1 | `state` |
| `klipper_component_info` | Gauge=1 | `component` |
| `klipper_component_failed_info` | Gauge=1 | `failed_component` |
| `klipper_moonraker_version_info` | Gauge=1 | `version` |
| `klipper_api_version_info` | Gauge=1 | `version` |

[Full reference →](./server-info)

### `network_stats`

| Metric | Type | Labels |
|--------|------|--------|
| `klipper_network_rx_bytes` | Counter | `interface` |
| `klipper_network_tx_bytes` | Counter | `interface` |
| `klipper_network_rx_packets` | Counter | `interface` |
| `klipper_network_tx_packets` | Counter | `interface` |
| `klipper_network_rx_errs` | Counter | `interface` |
| `klipper_network_tx_errs` | Counter | `interface` |
| `klipper_network_rx_drop` | Counter | `interface` |
| `klipper_network_tx_drop` | Counter | `interface` |
| `klipper_network_tx_bandwidth` | Gauge | `interface` |

[Full reference →](./network-stats)

### `system_info`

| Metric | Type | Labels |
|--------|------|--------|
| `klipper_system_cpu_count` | Gauge | |

[Full reference →](./system-info)

### `job_queue`

| Metric | Type | Labels |
|--------|------|--------|
| `klipper_job_queue_length` | Gauge | |
| `klipper_job_queue_state_info` | Gauge=1 | `state` |

[Full reference →](./job-queue)

### `directory_info`

| Metric | Type | Labels |
|--------|------|--------|
| `klipper_disk_usage_available` | Gauge | |
| `klipper_disk_usage_total` | Gauge | |
| `klipper_disk_usage_used` | Gauge | |

[Full reference →](./directory-info)

### `history`

| Metric | Type | Labels |
|--------|------|--------|
| `klipper_total_jobs` | Gauge | |
| `klipper_total_time` | Gauge | |
| `klipper_total_print_time` | Gauge | |
| `klipper_total_filament_used` | Gauge | |
| `klipper_longest_job` | Gauge | |
| `klipper_longest_print` | Gauge | |
| `klipper_current_print_first_layer_height` | Gauge | |
| `klipper_current_print_layer_height` | Gauge | |
| `klipper_current_print_object_height` | Gauge | |
| `klipper_current_print_total_duration` | Gauge | |

[Full reference →](./history)

### `printer_objects`

| Object | Metrics | Labels |
|--------|---------|--------|
| `controller_fan` | `klipper_controller_fan_rpm`, `klipper_controller_fan_speed` | `fan` |
| `display_status` | `klipper_print_gcode_progress` | |
| `extruder` | `klipper_extruder_power`, `klipper_extruder_pressure_advance`, `klipper_extruder_smooth_time`, `klipper_extruder_target`, `klipper_extruder_temperature` | |
| `fan` | `klipper_fan_rpm`, `klipper_fan_speed` | |
| `filament_sensor` | `klipper_filament_sensor_detected`, `klipper_filament_sensor_enabled` | `sensor` |
| `gcode` | `klipper_gcode_extrude_factor`, `klipper_gcode_position_e`, `klipper_gcode_position_x`, `klipper_gcode_position_y`, `klipper_gcode_position_z`, `klipper_gcode_speed_factor`, `klipper_gcode_speed` | |
| `generic_fan` | `klipper_generic_fan_rpm`, `klipper_generic_fan_speed` | `fan` |
| `heater_bed` | `klipper_heater_bed_power`, `klipper_heater_bed_target`, `klipper_heater_bed_temperature` | |
| `heater_generic` | `klipper_generic_heater_power`, `klipper_generic_heater_target`, `klipper_generic_heater_temperature` | `heater` |
| `idle_timeout` | `klipper_printing_time` | |
| `mcu` | `klipper_mcu_awake`, `klipper_mcu_task_avg`, `klipper_mcu_task_stddev`, `klipper_mcu_clock_frequency`, `klipper_mcu_invalid_bytes`, `klipper_mcu_read_bytes`, `klipper_mcu_ready_bytes`, `klipper_mcu_receive_seq`, `klipper_mcu_retransmit_bytes`, `klipper_mcu_retransmit_seq`, `klipper_mcu_rto`, `klipper_mcu_rttvar`, `klipper_mcu_send_seq`, `klipper_mcu_stalled_bytes`, `klipper_mcu_srtt`, `klipper_mcu_write_bytes` | `mcu` |
| `output_pin` | `klipper_output_pin_value` | `pin` |
| `print_stats` | `klipper_print_filament_used`, `klipper_print_total_duration`, `klipper_print_state_info{state="..."}`, `klipper_printing` | |
| `temperature_fan` | `klipper_temperature_fan_speed`, `klipper_temperature_fan_temperature`, `klipper_temperature_fan_target` | `fan` |
| `temperature_probe` | `klipper_temperature_probe_temperature`, `klipper_temperature_probe_measured_max_temp`, `klipper_temperature_probe_measured_min_temp`, `klipper_temperature_probe_estimated_expansion` | `probe` |
| `temperature_sensor` | `klipper_temperature_sensor_temperature`, `klipper_temperature_sensor_measured_max_temp`, `klipper_temperature_sensor_measured_min_temp` | `sensor` |
| `tmc_sensor` | `klipper_tmc_sensor_enabled`, `klipper_tmc_sensor_run_current`, `klipper_tmc_sensor_temperature` | `sensor` |
| `toolhead` | `klipper_toolhead_estimated_print_time`, `klipper_toolhead_max_accel`, `klipper_toolhead_max_accel_to_decel`, `klipper_toolhead_max_velocity`, `klipper_toolhead_print_time`, `klipper_toolhead_square_corner_velocity` | |
| `virtual_sdcard` | `klipper_print_file_position`, `klipper_print_file_progress` | |

[Full reference →](./printer-objects)

### `mmu`

MMU (Multi-Material Unit) metrics for Happy Hare. 50+ metrics covering:

| Category | Metrics |
|----------|---------|
| Basic state | `klipper_mmu_enabled`, `klipper_mmu_homed`, `klipper_mmu_num_gates`, `klipper_mmu_has_bypass`, `klipper_mmu_current_unit`, `klipper_mmu_current_tool`, `klipper_mmu_current_gate` |
| Print state | `klipper_mmu_print_state_info{state="..."}` |
| Action/Operation | `klipper_mmu_action_info{action="..."}`, `klipper_mmu_operation_info{operation="..."}` |
| Filament | `klipper_mmu_filament_loaded`, `klipper_mmu_filament_position_mm`, `klipper_mmu_filament_direction` |
| Toolchange | `klipper_mmu_toolchanges_total`, `klipper_mmu_last_tool`, `klipper_mmu_next_tool` |
| Encoder | `klipper_mmu_encoder_position_mm`, `klipper_mmu_encoder_headroom_mm`, `klipper_mmu_encoder_flow_rate_percent` |
| Per-gate | `klipper_mmu_gate_status{gate="..."}`, `klipper_mmu_gate_info{gate="...",material="...",color="..."}` |
| Sensors | `klipper_mmu_pre_gate_sensor_detected{gate="..."}` |
| Machine info | `klipper_mmu_machine_info{name="...",vendor="..."}` |
| Active filament | `klipper_mmu_active_filament_info{name="...",material="...",color="..."}` |
| Sync drive | `klipper_mmu_sync_drive_enabled`, `klipper_mmu_sync_feedback_state_info{state="..."}` |

[Full reference →](./mmu)

## Prometheus Metric Types

- **Gauge**: An instantaneous value that can go up or down (temperature, fan speed, queue length)
- **Counter**: A monotonically increasing value (uptime, total jobs)
- **Gauge=1 (Info)**: A gauge with value 1 and a label carrying the state value (e.g. `klipper_print_state_info{state="printing"} 1`)

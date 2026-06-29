Change Log
==========

unreleased
-------

- Add optional `cfs` module for the Creality Filament System (CFS) on K2-class printers, reading the native `box`, `filament_rack`, and `load_ai` Moonraker objects. Adds `klipper_cfs_enabled`, `klipper_cfs_auto_refill_enabled`, `klipper_cfs_filament_useup`, `klipper_cfs_state_info`, `klipper_cfs_active_unit`, `klipper_cfs_active_slot`, `klipper_cfs_active_slot_info`, `klipper_cfs_unit_state_info`, `klipper_cfs_unit_temperature_celsius`, `klipper_cfs_unit_humidity_percent`, `klipper_cfs_unit_info`, `klipper_cfs_slot_info`, `klipper_cfs_slot_remaining`, `klipper_cfs_rack_loaded_info`, `klipper_cfs_rack_velocity`, `klipper_cfs_ai_detection_enabled`, `klipper_cfs_ai_waste_detection_enabled`, `klipper_cfs_ai_max_probability`, `klipper_cfs_ai_normalized_area`, and `klipper_cfs_ai_command_info` metrics

v0.15.0
-------

- Add `klipper_printing` and `klipper_print_state_info` metrics. Address issue #46
- Add `klipper_temperature_fan_rpm` metric to `printer_objects` module
- Add `klipper_heater_fan_speed` and `klipper_heater_fan_rpm` metrics for `heater_fan` support in `printer_objects` module
- Add `klipper_toolhead_homed_axes_info` and `klipper_toolhead_stalls_total` metrics to `printer_objects` module
- Add `klipper_input_shaper_frequency_{x,y}`, `klipper_input_shaper_damping_ratio_{x,y}`, `klipper_input_shaper_type_info` metrics to `printer_objects` module
- Add `klipper_firmware_retract_length`, `klipper_firmware_retract_speed`, `klipper_firmware_unretract_extra_length`, `klipper_firmware_unretract_speed` metrics to `printer_objects` module
- Add `klipper_webhooks_state_info`, `klipper_pause_resume_is_paused`, `klipper_idle_timeout_state_info`, and `klipper_sdcard_active` metrics to `printer_objects` module
- Add `klipper_system_throttled_bits` and `klipper_system_throttled_flag_info` metrics to `process_stats` module
- Add `klipper_job_queue_state_info` metric to `job_queue` module
- Add `klipper_service_available`, `klipper_service_state_info`, and `klipper_service_sub_state_info` service state metrics to `system_info` module
- Add optional `device_power` module with `klipper_power_device_info`, `klipper_power_device_status`, and `klipper_power_device_state_info` metrics
- Add optional `spoolman` module with `klipper_spoolman_connected`, `klipper_spoolman_active_spool_id`, `klipper_spoolman_pending_reports`, `klipper_spoolman_spool_info`, `klipper_spoolman_remaining_weight`, `klipper_spoolman_used_weight`, `klipper_spoolman_remaining_length`, and `klipper_spoolman_used_length` metrics
- Add optional `server_info` module with `klipper_klippy_connected`, `klipper_klippy_state_info`, `klipper_component_info`, `klipper_component_failed_info`, `klipper_moonraker_version_info`, and `klipper_api_version_info` metrics
- Add optional `query_endstops` module with `klipper_endstop_triggered` metric
- Fix missing `temperature_probe` assignment in `printer_objects` UnmarshalJSON
- Fix missing `temperature_probe` query loop in `printer_objects` collector
- `server_info`, `query_endstops`, and `device_power` modules are now enabled by default
- Updated virtual test environment with additional metrics enabled and pre-built Grafana dashboards
- Add VitePress documentation site with metrics reference, guides, and local search

v0.14.0
-------

- Add support for `tmz_sensor` metrics for monitoring tmc stepper drivers like tmc2240. Thanks to @martijnvanduijneveldt #45
- Add support for `mmu` multi material unit metrics (Happy Hare). Thanks to @Alph4d0g #44
- Add local test environment with prometheus, grafana, virtual klipper and mainsail docker containers.
- Removed previously deprecated `temperature` metrics.
- Bumped supported golang version to 1.25

v0.13.0
-------

- Add support for `temperature_probe` metrics. Address issue #41
- Add support for `heater_generic` metrics. Thanks to @DavidvtWout #42

v0.12.0
-------

- Add support for multiple mcu's. Thanks to @Wulfsta #40
- Improved error logging for indexed results to address issues #34 and #35

v0.11.2
-------

- Fix regression in v0.11.1. Error in one collector module should not stop other collectors from attepting to run.
- Removed `.local` hostname in example config as container based setup doesn't work with mDNS.

v0.11.1
-------

- Refactor error handling to address #31

v0.11.0
-------

- Add support for filament sensors, controller fan, and generic fans. Thanks to @nmaggioni #28
- Get logging level from `LOGGING_LEVEL` env var if present. Thanks to @nmaggioni #29
- (Breaking change) Removed deprecated `-debug` and `-verbose` command line options.
  Use `-logging.level` option instead.

v0.10.3
-------

- Fix unmarshalling of total_jobs value in job history. Thanks to @jangrewe #27
- changed Dockerfile to use ENTRYPOINT instead of CMD

v0.10.2
-------

- Fixed Panic when the network in unreachable #24

v0.10.1
-------

- Fixed out of range errors for some metrics. Thanks to @hsmade #19

v0.10.0
-------

- Added metrics related to the current print (@danilodorgam #18)
  `klipper_current_print_*`, `klipper_gcode_position_*`

v0.9.0
------

- Added new MCU `klipper_mcu_*` metrics to `printer_objects` metric collection.
- Added new `-logging.level <level>` command line option to set specific log
  output level. The `-debug` and `-verbose` options have been deprecated and
  will be removed in a future release. Address #17.

v0.8.0
------

- Added option to set API Key for authentication in `prometheus.yml`, `-moonraker.apikey`
  command line option, or `MOONRAKER_APIKEY` environment variable. Fixes #15.
- Added `-verbose` option for trace level debug logging
- Breaking change: The `temperature` module is deprecated as it contains a subset
  of the metrics reported by the `printer_objects` module. Closes #2.

v0.7.1
------

- Added history data metrics including total print time or total filament used.
  Add the new `history` module in your `prometheus.yml` config. Thanks to @r4ptor #12

v0.7.0
------

- Fixes #11. Query custom temperature sensor, fan, and output pin config separatelly
  for each configured klipper host
- Fixes #10. Use labels for temperature sensor, fan, and output pin metrics
- Breaking change: `printer_objects` `klipper_temperature_sensor_*` metrics
  renamed, now uses labels for each sensor instead of separate metrics
- Breaking change: `printer_objects` `klipper_temperature_fan_*` metrics
  renamed now uses labels for each fan instead of separate metrics
- Breaking change: `printer_objects` `klipper_output_pin_*` metrics renamed, now
  uses labels for each output pin instead for separate metrics
- Breaking change: `network_stats` `klipper_network_*` metrics renamed, now
  uses labels for each network interface instead of separate metrics

v0.6.2
------

- Fixes #9. Change TimeInQueue type from into to float64

v0.6.1
------

- Fixes #8. Invalid metric for sensors with unsupported characters in the name

v0.6.0
------

- Added support for `temperature_sensor` metrics to `printer_objects` collector. Fixes #3
- Added support for `temperature_fan` metrics to `printer_objects` collector. Fixes #4
- Added support for `output_pin` metrics to `printer_objects` collector. Fixes #5
- Added [example](./example/) docker deployment for grafana, prometheus and the klipper-exporter

v0.5.1
------

- Added `Dockerfile` and Docker usage instructions. Fixes [#6](https://github.com/scross01/prometheus-klipper-exporter/issues/6)
- Fixes issue with linux builds [#1](https://github.com/scross01/prometheus-klipper-exporter/issues/1)
- Fixes typo in metric descriptions

v0.5.0
------

- Add additional `printer_object` metrics from idle_timeout, virtual_sdcard, print_stats, and display_staus
- Changes some metric types from Gauge to Counter
- Fixes heater bed metric collection

v0.4.0
------

- Add `printer_object` metrics for gcode_move, toolheat, extruder, heater_bed, and fan.
- Add `temperature` metric collection

v0.3.0
------

- Separate metrics into optional modules
- Add network stats for all network interfaces
- Update logging
- Fixes range exception for large gauges on 32-bits rpi OS

v0.2.0
------

- Adds some system metrics from /machine/system_info
- Adds some Disk Storage mertics from /server/files/directory
- Add Job Queue metric from /server/job_queue/status queue length
- Added build targets for different platforms

v0.1.1
------

- Fixes crash when moonraker API is offline
- Remove foo bar test metrics

v0.1.0
------

- Initial version with support for `/machine/proc_stats`

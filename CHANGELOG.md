Change Log
==========

v0.12.0
-------

- Add support for multiple mcu's. Thanks to @Wulfsta #40.
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

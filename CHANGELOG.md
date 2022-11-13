Change Log
==========

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

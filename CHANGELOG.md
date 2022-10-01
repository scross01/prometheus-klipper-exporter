Change Log
==========

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

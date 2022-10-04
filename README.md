Prometheus Exporter for Klipper
===============================

Initial implmentation of a Prometheus exporter for Klipper to capture operational
metrics. This is a very rough first implementation and subject to potentially
significant changes.

Implementation is based on the [Multi-Target Exporter Pattern](https://prometheus.io/docs/guides/multi-target-exporter/)
to enabled a single exporter to collect metrics from multiple Klipper instances.
Metrics for the exporter itself are served from the `/metrics` endpoint and Klipper
metrics are serviced from the `/probe` endpoint with a specified `target`.

Usage
-----

To start the Prometheus Klipper Exporter from the command line

```sh
$ prometheus-klipper-exporter
INFO[0000] Beginning to serve on port :9101             
```

Then add a Klipper job to the Prometheus configuration file `/etc/prometheus/prometheus.yml`

```yaml
scrape_configs:

  - job_name: "klipper"
    scrape_interval: 5s
    metrics_path: /probe
    static_configs:
      - targets: [ 'klipper.local:7125' ]
    params:
      modules: [ "process_stats", "job_queue", "system_info", "network_stats", "directory_info", "temperature", "printer_objects" ]
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: klipper-exporter.local:9101

  # optional exporter metrics
  - job_name: "klipper-exporter"
    scrape_interval: 5s
    metrics_path: /metrics
    static_configs:
      - targets: [ 'klipper-exporter.local:9101' ]
```

The exporter can be run on the host running Klipper, or on a separate machine.
Replace `klipper.local` with the hostname or IP address of the Klipper host,
and replace `klipper-exporter.local` with the hostname or IP address of the host
runnging `prometheus-klipper-exporter`.

To monitor multiple Klipper instances add multiple entries to the
`static_config`.`targets` for the `klipper` job. e.g.

```yaml
    ...
    static_configs:
      - targets: [ 'klipper1.local:7125', 'klipper2.local:7125 ]
    ...
```

Build
-----

```sh
make build
```

Modules
-------

You can configure different sets of metrics to be collected by including the
`modules` parameter in the `prometheus.yml` configuration file.

```yaml
    ...
    params:
      modules: [ "process_stats", "job_queue", "system_info" ]
    ...
```

If the modules params are omitted then only the default metrics are collected. Each
group of metrics is queried from a different Moonraker API endpoint.

| module | default | metrics |
|--------|---------|---------|
| `process_stats` | x | `klipper_moonraker_cpu_usage`<br/>`klipper_moonraker_memory_kb`<br/>`klipper_moonraker_websocket_connections`<br/>`klipper_system_cpu`<br/>`klipper_system_cpu_temp`<br/>`klipper_system_memory_available`<br/>`klipper_system_memory_total`<br/>`klipper_system_memory_used`<br/>`klipper_system_uptime`<br/> |
| `network_stats` |   | Per interface (e.g. `lo`, `wlan`):<br/>`klipper_network_*_tx_bandwidth`<br/>`klipper_network_*_rx_bytes`<br/>`klipper_network_*_tx_bytes`<br/>`klipper_network_*_rx_drop`<br/>`klipper_network_*_tx_drop`<br/>`klipper_network_*_rx_errs`<br/>`klipper_network_*_tx_errs`<br/>`klipper_network_*_rx_packets`<br/>`klipper_network_*_tx_packets` |
| `job_queue` | x | `klipper_job_queue_length` |
| `system_info` | x | `klipper_system_cpu_count` |
| `directory_info` | | `klipper_disk_usage_available`<br/>`klipper_disk_usage_total`<br/>`klipper_disk_usage_used` |
| `temperature` | | `klipper_extruder_power`<br/>`klipper_extruder_target`<br/>`klipper_extruder_temperature`<br/>`klipper_heater_bed_power`<br/>`klipper_heater_bed_target`<br/>`klipper_heater_bed_temperature`<br/>`klipper_temperature_sensor_*_temperature` |
| `printer_objects` | | `klipper_extruder_power`<br/>`klipper_extruder_pressure_advance`<br/>`klipper_extruder_smooth_time`<br/>`klipper_extruder_target`<br/>`klipper_extruder_temperature`<br/>`klipper_fan_rpm`<br/>`klipper_fan_speed`<br/>`klipper_gcode_extrude_factor`<br/>`klipper_gcode_speed_factor`<br/>`klipper_gcode_speed`<br/>`klipper_heater_bed_power`<br/>`klipper_heater_bed_target`<br/>`klipper_heater_bed_temperature`<br/>`klipper_printing_time`<br/>`klipper_print_filament_used`<br/>`klipper_print_file_position`<br/>`klipper_print_file_progress`<br/>`klipper_print_gcode_progress`<br/>`klipper_print_total_duration`<br/>`klipper_toolhead_estimated_print_time`<br/>`klipper_toolhead_max_accel_to_decel`<br/>`klipper_toolhead_max_accel`<br/>`klipper_toolhead_max_velocity`<br/>`klipper_toolhead_print_time`<br/>`klipper_toolhead_square_corner_velocity` |

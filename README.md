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
      modules: [
        "process_stats",
        "job_queue",
        "system_info",
        "network_stats",
        "directory_info",
        "printer_objects",
        "history",
      ]
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
$ make build
```

Installation
------------

You will typically want to install the exporter binary on a host that will be
constantly running, either the Klipper host iteself, or a separate server, and
ensure that the process restarts on system restart.

### systemd

Example installation on Raspberry Pi, using systemd to run the exporter.

```sh
$ ssh pi@klipper.local
[klipper]$ mkdir /home/pi/klipper-exporter
[klipper]$ exit

$ scp prometheus-klipper-exporter pi@klipper.local:/home/pi/klipper-exporter
$ scp klipper-exporter.service pi@klipper.local:/home/pi/

$ ssh pi@klipper.local
[klipper]$ sudo mv klipper-exporter.service /etc/systemd/system/
[klipper]$ sudo systemctl daemon-reload
[klipper]$ sudo systemctl enable klipper-exporter.service
[klipper]$ sudo systemctl start klipper-exporter.service
[klipper]$ sudo systemctl status klipper-exporter.service
[klipper]$ exit
```

### docker

To run the exporter as a docker container.

```sh
$ docker run -d -p 9101:9101 ghcr.io/scross01/prometheus-klipper-exporter:latest
```

See the [example/README.md](example/README.md) for a complete example running
Prometheus, Grafana, and the klipper-exporter in Docker using docker compose.

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
| `network_stats` |   | `klipper_network_tx_bandwidth{interface="`*interface*`"}`<br/>`klipper_network_rx_bytes{interface="`*interface*`"}`<br/>`klipper_network_tx_bytes{interface="`*interface*`"}`<br/>`klipper_network_rx_drop{interface="`*interface*`"}`<br/>`klipper_network_tx_drop{interface="`*interface*`"}`<br/>`klipper_network_rx_errs{interface="`*interface*`"}`<br/>`klipper_network_tx_errs{interface="`*interface*`"}`<br/>`klipper_network_rx_packets{interface="`*interface*`"}`<br/>`klipper_network_tx_packets{interface="`*interface*`"}`<br/> |
| `job_queue` | x | `klipper_job_queue_length` |
| `system_info` | x | `klipper_system_cpu_count` |
| `directory_info` | | `klipper_disk_usage_available`<br/>`klipper_disk_usage_total`<br/>`klipper_disk_usage_used` |
| `printer_objects` | | `klipper_extruder_power`<br/>`klipper_extruder_pressure_advance`<br/>`klipper_extruder_smooth_time`<br/>`klipper_extruder_target`<br/>`klipper_extruder_temperature`<br/>`klipper_fan_rpm`<br/>`klipper_fan_speed`<br/>`klipper_gcode_extrude_factor`<br/>`klipper_gcode_speed_factor`<br/>`klipper_gcode_speed`<br/>`klipper_heater_bed_power`<br/>`klipper_heater_bed_target`<br/>`klipper_heater_bed_temperature`<br/>`klipper_mcu_awake`<br/>`klipper_mcu_clock_frequency`<br/>`klipper_mcu_invalid_bytes`<br/>`klipper_mcu_read_bytes`<br/>`klipper_mcu_ready_bytes`<br/>`klipper_mcu_receive_seq`<br/>`klipper_mcu_retransmit_bytes`<br/>`klipper_mcu_retransmit_seq`<br/>`klipper_mcu_rto`<br/>`klipper_mcu_rttvar`<br/>`klipper_mcu_send_seq`<br/>`klipper_mcu_stalled_bytes`<br/>`klipper_mcu_srtt`<br/>`klipper_mcu_write_bytes`<br/>`klipper_output_pin_value{pin="`*pin*`"}`<br/>`klipper_printing_time`<br/>`klipper_print_filament_used`<br/>`klipper_print_file_position`<br/>`klipper_print_file_progress`<br/>`klipper_print_gcode_progress`<br/>`klipper_print_total_duration`<br/>`klipper_temperature_fan_speed{fan="`*fan*`"}`<br/>`klipper_temperature_fan_temperature{fan="`*fan*`"}`<br/>`klipper_temperature_fan_target{fan="`*fan*`"}`<br/>`klipper_temperature_sensor_temperature{sensor="`*sensor*`"}`<br/>`klipper_temperature_sensor_measured_max_temp{sensor="`*sensor*`"}`<br/>`klipper_temperature_sensor_measured_min_temp{sensor="`*sensor*`"}`<br/>`klipper_toolhead_estimated_print_time`<br/>`klipper_toolhead_max_accel_to_decel`<br/>`klipper_toolhead_max_accel`<br/>`klipper_toolhead_max_velocity`<br/>`klipper_toolhead_print_time`<br/>`klipper_toolhead_square_corner_velocity` |
| `history` | | `klipper_total_jobs`<br/>`klipper_total_time`<br/>`klipper_total_print_time`<br/>`klipper_total_filament_used`<br/>`klipper_longest_job`<br/>`klipper_longest_print`<br/>

Authentication
--------------

### Trusted Client

The simplest deployment option is to run the Klipper Exporter on a host that is in
the Moonraker trusted clients configuration.  This is typically configured by default
to include all hosts in the local network. If you have a more restrictive configuration
then add the host to the `moonraker.conf` [`[authorization]`](https://moonraker.readthedocs.io/en/latest/configuration/#authorization)
configuration section.

```yaml
# moonraker.conf

[authorization]
trusted_clients:
  klipper-exporter.local
  ...
```

### API Key Authentication

Untrusted clients must use an API key to access Moonraker's HTTP APIs. To fetch
the current API key run the following on the Klipper host:

```sh
$ cd ~/moonraker/scripts
$ ./fetch-apikey.sh
abcdef01234567890123456789012345
```

The API key can be set in one of three ways, from the scrape job configuraion in
`prometheus.yml`, using the `-moonraker-apikey` command line argument, or
setting the `MOONRAKER_APIKEY` environment variable.

#### Environment variable

Set in the `MOONRAKER_APIKEY` environment variable.

```sh
$ export MOONRAKER_APIKEY='abcdef01234567890123456789012345'
$ prometheus-klipper-exporter
```

#### Command line argument

Set on the klipper exporter command line using `-moonraker.apikey` option.

```sh
$ prometheus-klipper-exporter -moonraker.apikey='abcdef01234567890123456789012345'
```

#### Prometheus scrape configuration

Add the API key to the `prometheus.yml` scrape config, Add `authorization`
configuration with the type set to `APIKEY`.  The key can either to set directly
in the config or referenced from file.

```yaml
  - job_name: "klipper"
    ...
    authorization:
      type: APIKEY
      credentials: 'abcdef01234567890123456789012345'
      # credentials_file: /path/to/private/apikey.txt
    ...
```

Only one API key can be set for each job.  If you have multiple klipper hosts with
different API keys, create a separate job for each host.

Command line options
--------------------

`-help`

  Display the command line help.

`-logging.level <level>`

  Set the logging output verbosity to one of `Trace`, `Debug`, `Info`,
  `Warning`, `Error`, `Fatal` and `Panic`. Default level is `Info` which will
  log anything that is info level or above (warning, error, fatal, panic).

`-moonraker.apikey <string>`

  Set the API Key to authenticate with the Klipper APIs.
  See [API Key Authentication](#api-key-authentication)

`-web.listen-address [<ip_address>]:<port>`

  Address on which to expose metrics and web interface. Default is `:9101`
  which will listen on port `9101` on all interfaces, which is the equiviment
  of `0.0.0.0:9101`.  Include the IP address to limit to listening on a specific
  interface, e.g. `192.168.1.99:7070`.

⚠️ History of breaking changes
-----------------------------

### Upgrading to v0.8.0

`v0.8.0` deprecates the `tempurature` module option which contains a subset of
the metrics reported by the `printer_objects`. If you where using the
`tempurature` module then switch the configuration to use `printer_objects` instead.

### Upgrading to v0.7.0

The `v0.7.0` release introduces several metric changes that will break any
grafana charts that have previously been defined using the old metric names from
v0.6.x or earlier.

`v0.7.0` now uses labels for network interfaces, temperature sensors,
temperature fans, and output pins, rather than defining separate metrics
for each unique entity.

These changes affect the following metrics groups

- `klipper_network_`*
- `klipper_temperature_sensor_`*
- `klipper_temperature_fan_`*
- `klipper_output_pin_`*

For example:

- `klipper_network_`**<code>wlan0</code>**`_rx_bytes` becomes `klipper_network_rx_bytes{interface="`**<code>wlan0</code>**`"}`
- `klipper_temperature_sensor_`**<code>mtu</code>**`_temperature` becomes `klipper_temperature_sensor_temperature{sensor="`**<code>mtu</code>**`"}`
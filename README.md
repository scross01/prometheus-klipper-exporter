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

To monitor multiple Klipper instances add mutiple entries to the
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

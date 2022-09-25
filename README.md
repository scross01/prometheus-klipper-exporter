Prometheus Exporter for Klipper
===============================

Initial implmentation of a Prometheus exporter for Klipper to capture operational metrics. This is a very rough first implementation and subject to potentially significant changes.

Implementation is based on the [Multi-Target Exporter Pattern](https://prometheus.io/docs/guides/multi-target-exporter/) to enabled a single exporter to collect metrics from multiple Klipper instances. Metrics for the exporter itself are served from the `/metrics` endpoint and Klipper metrics are serviced from the `/probe` endpoint with a specified `target`.

Usage
-----

To start the Prometheus exporter from the command line.

```sh
$ prometheus-klipper-exporter
```

Then add a klipper job to the Prometheus configuration file `/etc/prometheus/prometheus.yml`

```yaml
scrape_configs:

  - job_name: "klipper"
    scrape_interval: 5s
    metrics_path: /probe
    params:
      target: [ "klipper.local:7125" ]
    static_configs:
      - targets: [ 'klipper.local:9101' ]

  # optional exporter metrics
  - job_name: "klipper-exporter"
    scrape_interval: 5s
    metrics_path: /metrics
    static_configs:
      - targets: [ 'klipper.local:9101' ]
```

The klipper exporter can run on your Klipper host, or on a separate server. The `static_configs`.`target` needs to be set to the host and port of the  `prometheus-klipper-exporter` and `params`.`target` is the hostname and port of the the Klipper installation Moonraker API.

Build
-----

```sh
make build
```

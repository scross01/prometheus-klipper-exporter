---
layout: home

hero:
  name: Prometheus Klipper Exporter
  tagline: Capture operational metrics from Klipper 3D printer firmware
  actions:
    - theme: brand
      text: Get Started
      link: /guide/
    - theme: alt
      text: Metrics Reference
      link: /metrics/
    - theme: alt
      text: GitHub
      link: https://github.com/scross01/prometheus-klipper-exporter

features:
  - title: Multi-Target Exporter
    details: Collect metrics from multiple Klipper instances using a single exporter following the Prometheus multi-target exporter pattern.
  - title: Modular Collection
    details: Enable only the metric groups you need with a configurable modules system. Each module maps to specific Moonraker API endpoints.
  - title: Fleet Monitoring
    details: Monitor temperature, fan speed, print progress, system resources, and more across all your printers in one place.
  - title: Prometheus Native
    details: Exposes metrics in standard Prometheus format with full label support for dimensional monitoring and alerting.
---

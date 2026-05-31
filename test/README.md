Virtual Test Environment
=======================

```shell
docker compose up -d
```

Klipper/Moonraker: http://localhost:7125
Mainsail: http://localhost:8080 
Prometheus: http://localhost:9090
Grafana: http://localhost:3000

Grafana Dashboards
------------------

Example dashboards are auto-provisioned on startup in a **Klipper** folder:

- **Klipper System** — CPU, memory, network, disk, Moonraker process stats
- **Klipper Temperatures** — Extruder, bed, sensors, fans, probes, heaters
- **Klipper Print Status** — Print progress, g-code position, history
- **Klipper Hardware** — MCU stats, fan speeds/RPMs, output pins, TMC drivers
- **Klipper MMU** — Multi-Material Unit state, gate status, encoder

Select `job=klipper` and `instance=klipper:7125` in the dashboard selectors.

Troubleshooting
---------------

If Mainsail is unabled to connect to the virtual printer it is likely due to
the browsers automatic HTTP to HTTPS upgrade. You will need to disable HTTPS
Upgrade for the site.

https://support.mozilla.org/en-US/kb/https-only-prefs?utm_source=mozilla&utm_medium=firefox-console-errors&utm_campaign=default&as=u

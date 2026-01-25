Virtual Test Environment
========================

```shell
docker compose up -d
```

Klipper/Moonraker: http://localhost:7125
Mainsail: http://localhost:8080 
Prometheus: http://localhost:9090
Grafana: http://localhost:3000

Troubleshooting
---------------

If Mainsail is unabled to connect to the virtual printer it is likely due to
the browsers automatic HTTP to HTTPS upgrade. You will need to disable HTTPS
Upgrade for the site.

https://support.mozilla.org/en-US/kb/https-only-prefs?utm_source=mozilla&utm_medium=firefox-console-errors&utm_campaign=default&as=u

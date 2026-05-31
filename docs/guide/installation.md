# Installation

You can install the exporter binary directly on the Klipper host or on a separate
monitoring server. The exporter does not need to run on the Klipper host.

## Build from source

```sh
$ git clone https://github.com/scross01/prometheus-klipper-exporter
$ cd prometheus-klipper-exporter
$ make build
```

The binary is built as `prometheus-klipper-exporter` in the project root.

## systemd (Raspberry Pi / Linux)

Example installation on a Raspberry Pi running Klipper, using systemd to manage
the exporter process.

```sh
$ ssh pi@klipper-host
[klipper]$ mkdir /home/pi/klipper-exporter
[klipper]$ exit

$ scp prometheus-klipper-exporter pi@klipper-host:/home/pi/klipper-exporter
$ scp klipper-exporter.service pi@klipper-host:/home/pi/

$ ssh pi@klipper-host
[klipper]$ sudo mv klipper-exporter.service /etc/systemd/system/
[klipper]$ sudo systemctl daemon-reload
[klipper]$ sudo systemctl enable klipper-exporter.service
[klipper]$ sudo systemctl start klipper-exporter.service
[klipper]$ sudo systemctl status klipper-exporter.service
```

## Docker

```sh
$ docker run -d -p 9101:9101 ghcr.io/scross01/prometheus-klipper-exporter:latest
```

## Docker Compose

See the [example docker-compose stack](https://github.com/scross01/prometheus-klipper-exporter/tree/main/example)
for a complete setup with Prometheus and Grafana.

```yaml
services:
  klipper-exporter:
    image: ghcr.io/scross01/prometheus-klipper-exporter:latest
    container_name: klipper-exporter
    restart: unless-stopped
    ports:
      - "9101:9101"
```

> **Note:** If the container cannot resolve local Klipper hostnames, add a DNS
> setting to the compose file:
> ```yaml
>     dns:
>       - 192.168.1.1
> ```

## Building for other platforms

The project supports cross-compilation for multiple platforms:

```sh
$ make release
```

This produces binaries in `build/release-<version>/` for:
- `linux/amd64`, `linux/arm64`, `linux/arm/v7` (Raspberry Pi)
- `darwin/amd64`, `darwin/arm64` (macOS)
- `windows/amd64` (Windows)

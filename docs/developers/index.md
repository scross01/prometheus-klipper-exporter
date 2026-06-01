# Developers

This page provides detailed information for developers working on the Prometheus
Klipper Exporter.

## Project Structure

```
.
├── main.go                         # HTTP server, routing, CLI flags
├── collector/
│   ├── collector.go                # Prometheus Collector interface, shared utilities
│   ├── device_power.go            # /machine/device_power (power device status)
│   ├── directory_info.go           # /server/files/directory
│   ├── history.go                  # /server/history/totals
│   ├── job_queue.go                # /server/job_queue/status
│   ├── network_stats.go            # /machine/proc_stats (network interfaces)
│   ├── printer_object.go           # /printer/objects/query
│   ├── process_stats.go            # /machine/proc_stats (CPU/memory)
│   ├── spoolman.go                # POST /server/spoolman/proxy → GET /v1/spool (Spoolman filament spools)
│   ├── system_info.go              # /machine/system_info (CPU count and service states)
│   └── mmu.go                      # /printer/objects/query (MMU objects)
├── test/
│   └── README.md                   # Quick start for test env
│   ├── docker-compose.yml          # Local test environment
│   ├── printer_data/               # Virtual Klipper printer config
│   ├── prometheus.yml              # Prometheus scrape config for local dev
├── docs/                           # VitePress documentation site
├── example/                        # Docker deployment example
└── Makefile                        # Build, fmt, test, release targets
```

### Key Architectural Patterns

- **Multi-Target Exporter**: A single exporter instance scrapes multiple Klipper
  hosts using the `/probe?target=<host>` endpoint
- **Collector Interface**: Each module implements `prometheus.Collector`
  (`Describe()` + `Collect()`)
- **Module Gating**: Features are enabled via `slices.Contains(c.modules, "name")`
  guards in `Collect()`
- **API Key Priority**: Header > CLI flag (`-moonraker.apikey`) > Environment
  variable (`MOONRAKER_APIKEY`)

## Building and Testing

### Prerequisites

- Go 1.25+
- Make

### Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the exporter binary |
| `make run` | Run the exporter locally |
| `make fmt` | Format Go code |
| `make test` | Run all tests |
| `make release` | Cross-compile for all platforms |

### Cross-Compilation

The `make release` target builds for Raspberry Pi (ARM), Linux (AMD64), macOS
(AMD64/ARM64), and Windows (AMD64) using `GOOS`/`GOARCH` environment variables.

### Running Tests

```sh
make test
```

Test files live in `tests/` and follow standard Go testing patterns.

## Virtual Printer Test Environment

The `test/` directory contains a Docker Compose-based test environment. It
supports two virtual printer images that you can choose between in
`test/docker-compose.yml`:

| Option | Image | Description |
|--------|-------|-------------|
| **A** _(default)_ | `ghcr.io/pedrolamas/docker-klipper-simulavr` | Full MCU simulation via `simulavr`. Supports `SIMULAVR_PACING_RATE` to prevent "Timer too close" errors. Config path: `/printer/printer_data` |
| **B** | `ghcr.io/mainsail-crew/virtual-klipper-printer` | Lighter API-level simulation. No MCU simulation. Config path: `/home/printer/printer_data` |

To switch between them, uncomment your preferred option and comment out the
other in `test/docker-compose.yml`.

### Starting the Environment

```sh
docker compose up -d --build  # from test/ directory
```

### Services

| Service | URL |
|---------|-----|
| Klipper/Moonraker | `http://localhost:7125` |
| Mainsail | `http://localhost:8080` |
| Prometheus | `http://localhost:9090` |
| Grafana | `http://localhost:3000` |

### Tuning `SIMULAVR_PACING_RATE`

When using Option A (`docker-klipper-simulavr`), the `SIMULAVR_PACING_RATE`
environment variable controls how fast the simulated MCU runs. If you see
`MCU 'mcu' shutdown: Timer too close` in the container logs:

1. **Reduce** the value in `test/docker-compose.yml` (e.g. `0.1` → `0.05` → `0.01`)
2. Restart the container: `docker compose restart virtual-klipper`
3. A lower value gives the host CPU more headroom by slowing the simulation
4. Values above `0.2` are not recommended — they can starve the simulator and
   make the timer issue worse
5. The default of `0.1` works on most modern hardware; adjust based on your
   system's CPU load

### Virtual Printer Configuration

The virtual printer config lives in `test/printer_data/config/`. The
`printer.cfg` includes addon configs from `test/printer_data/config/addons/`.

#### Pin Assignments

The virtual MCU is an AVR atmega644p with the following available pins:

- **PORTA**: PA0, PA1, PA2, PA3, PA4, PA5, PA6, PA7
- **PORTB**: PB0, PB1, PB2, PB3, PB4, PB5, PB6, PB7
- **PORTC**: PC0, PC1, PC2, PC3, PC4, PC5, PC6, PC7
- **PORTD**: PD0, PD1, PD2, PD3, PD4, PD5, PD6, PD7

#### Addon Configs

| Addon File | Sections Defined | Pins Used |
|------------|-----------------|-----------|
| `basic_cartesian_kinematics.cfg` | `stepper_x`, `stepper_y`, `stepper_z`, `extruder` | step/dir pins |
| `basic_macros.cfg` | G-code macros | — |
| `single_extruder.cfg` | `extruder` | heater, sensor pins |
| `heater_bed.cfg` | `heater_bed` | heater, sensor pins |
| `temp_sensors.cfg` | `temperature_sensor`, `temperature_fan` | PA1, PA4, PD2, PD3 |
| `miscellaneous.cfg` | `fan`, `heater_fan`, `controller_fan`, `filament_motion_sensor`, `output_pin` | PB4, PB5, PB6, PC0, PC1 |
| `custom_features.cfg` | `temperature_probe`, `heater_generic` | PA0, PA2, PA3 |
| `input_shaper.cfg` | `input_shaper` | — (no pins required) |
| `timelapse.cfg` | Moonraker timelapse | — |

#### Prometheus Scrape Config

The test environment's `prometheus.yml` scrapes these modules:

```yaml
params:
  modules:
    - process_stats
    - network_stats
    - system_info
    - job_queue
    - directory_info
    - printer_objects
    - history
    - device_power
```

All metrics from these modules are available at `http://localhost:9101/probe?target=virtual-klipper:7125`.

### Grafana Dashboards

The test environment includes auto-provisioned example Grafana dashboards,
loaded from `test/grafana/provisioning/dashboards/`. They are available at
`http://localhost:3000` under the **Klipper** folder:

| Dashboard | Focus | Key Metrics |
|-----------|-------|-------------|
| **Klipper System** | System health | CPU, memory, uptime, network, disk, job queue, Moonraker process, service states |
| **Klipper Temperatures** | Temperature monitoring | Extruder, bed, sensors, temperature fans, probes, generic heaters |
| **Klipper Print Status** | Print progress & history | G-code progress, file position, filament used, timeline, history stats |
| **Klipper Hardware** | MCU, fans, pins, TMC, power devices | MCU task/RTT/I/O, fan speeds/RPMs, output pins, filament sensors, TMC drivers, power device status |
| **Klipper MMU** | Multi-Material Unit | Gate/tool state, encoder data, filament status, toolchange tracking |

The dashboards use `job` and `instance` template variables. For the test
environment select `job=klipper` and `instance=virtual-klipper:7125`.

Each dashboard can also be imported manually into another Grafana instance from
the JSON files in `test/grafana/provisioning/dashboards/`. The JSON uses a
`DS_PROMETHEUS` datasource input variable — you will be prompted to map it
during import.

### Adding New Config Sections

When adding a new Klipper config section to exercise exporter code:

1. **Check pin conflicts**: Ensure the pin isn't already used by another addon.
   Available pins are listed above.
2. **Add or modify an addon file**: Create a new `.cfg` in `addons/` or modify
   an existing one.
3. **Include it in `printer.cfg`**: Add an `[include addons/your_file.cfg]` line.
4. **Restart the container**: The virtual printer will reload config on restart.

### Known Issues

**MCU `Timer too close` shutdown in simulavr**

When using Option A (`docker-klipper-simulavr`), the `simulavr` MCU emulator
can trigger `MCU shutdown: Timer too close` if the host is under load. This is
mitigated by setting `SIMULAVR_PACING_RATE` to a lower value — see the
[Tuning section](#tuning-simulavrpacingrate) above.

When using Option B (`virtual-klipper-printer`), the MCU shutdown is
consistently triggered ~7 seconds after Klippy reaches the `ready` state by
Moonraker's `objects/query` request. The old workaround was to restart the
full stack repeatedly until a stable cycle occurred.

Symptoms (either option):
- Klippy cycles through ready → shutdown → auto-restart every ~30-60 seconds
- Moonraker reports `klippy_state=shutdown` even while Klippy produces Stats
- The UDS socket (`klippy.sock`) exists but does not respond to API requests
- Metrics continue to flow through the exporter because Moonraker caches last
  known values and serves them regardless of Klippy's state

Impact on development:
- Metrics are served and scraped correctly during both `ready` and `shutdown`
  states, so the exporter and dashboards remain functional
- Printer-object metrics (temperatures, fans, sensors) reflect the last cached
  values, not live readings, during a shutdown cycle
- Switching to Option A with a properly tuned `SIMULAVR_PACING_RATE` can
  keep Klippy in `ready` state indefinitely

## Collector Implementation Guide

### Adding a New Module

1. Create a new file in `collector/` with:
   - A `collect*()` method that fetches data and emits metrics
   - Helper types for JSON response unmarshalling
   - A `fetchMoonraker*()` function for the API call

2. Register the module in `collector.go`'s `Collect()` method:
   ```go
   if slices.Contains(c.modules, "your_module") {
       c.collectYourModule(ch, target, apikey)
   }
   ```

3. If the module should be enabled by default, add it to the default modules
   list in `main.go`.

### Metric Naming Conventions

- **Prefix**: `klipper_*`
- **Case**: snake_case
- **Suffix conventions**:
  - `_info` — labeled state gauges for enumerated values (e.g. `klipper_print_state_info{state="printing"}`)
  - `_total` — counters
  - `_celsius`, `_mm`, `_seconds` — unit suffixes
- **Do not** expose arbitrary strings (error messages, filenames) as labels

### Shared Utilities (in `collector.go`)

| Function | Purpose |
|----------|---------|
| `GetValidLabelName()` | Converts hyphens to underscores, strips invalid characters |
| `boolToFloat64()` | Converts `bool` to `0.0`/`1.0` for Prometheus |
| `emitStateInfoMetric()` | Emits a `_info` metric for string states with known values |

### Error Handling

Log the error with `log.Error(err)` and return early. An error in one module
should not prevent other modules from collecting.

## Documentation Site

The documentation site uses [VitePress](https://vitepress.dev/) and is deployed
to GitHub Pages from the `main` branch.

### Previewing Locally

```sh
cd docs
npm install
npm run dev     # hot-reload dev server
npm run build   # production build
npm run preview # preview production build
```

### Adding a Docs Page

1. Create the `.md` file in the appropriate `docs/` subdirectory
2. Register it in `docs/.vitepress/config.js` in the relevant sidebar section
3. Verify the build with `npm run build`

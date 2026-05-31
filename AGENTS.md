# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Project Overview
- **Language**: Go 1.25
- **Framework**: Prometheus exporter for Klipper 3D printer firmware
- **Build System**: Makefile-based with cross-compilation support

## Build/Run Commands
- `make build` - Build the exporter
- `make run` - Run the exporter locally
- `make fmt` - Format Go code (runs `go fmt` in both root and collector directories)
- `make release` - Build for all platforms (RPi, Linux, macOS, Windows)
- `make test` - Run tests
- `cd docs && npm run dev` - Start VitePress docs dev server (hot-reload)
- `cd docs && npm run build` - Build docs site for production
- `go run .` - Direct execution

## Architecture
- **Main Entry**: `main.go` handles HTTP server and routing
- **Collector Pattern**: `collector/collector.go` implements Prometheus Collector interface
- **Module System**: Each collector file (process_stats.go, job_queue.go, etc.) handles specific Klipper API endpoints
- **API Key Priority**: Header > CLI arg > Environment variable
- **Documentation Site**: `docs/` is a VitePress site, deployed to GitHub Pages from `main`

## Non-Obvious Patterns
- **Metric Name Sanitization**: `GetValidLabelName()` in collector.go converts hyphens to underscores and strips invalid characters
- **Boolean Conversion**: `boolToFloat64()` converts booleans to 0/1 for Prometheus metrics
- **Module Registration**: Collector checks `slices.Contains(c.modules, "module_name")` to enable/disable features
- **API Response Handling**: Each collector file has `fetchMoonraker*` functions that handle HTTP requests and JSON parsing

## Code Style
- **Markdown**: Follows `.markdownlint.json` rules (setext_with_atx headers, line length)
- **Go**: Standard Go formatting with `go fmt`
- **Naming**: Prometheus metric names use `klipper_*` prefix with snake_case
- **Enumerated State Fields**: Use `emitStateInfoMetric()` for string states with known values (e.g. `klipper_print_state_info{state="printing"}`). Do not expose arbitrary strings (error messages, filenames) as labels.

## Critical Gotchas
- **Cross-Compilation**: Uses `GOOS`/`GOARCH` environment variables for platform-specific builds
- **Collector Pattern**: Must implement both `Describe()` and `Collect()` methods
- **Error Handling**: Returns early on errors but logs them first
- **Context Usage**: Collector uses context for cancellation and timeouts
- **`network_stats` dependency**: The `network_stats` module shares the `/machine/proc_stats` endpoint and is gated alongside `process_stats` in `Collect()` — it is not a standalone fetch
- **Module naming in queries**: New collector files must be registered in `collector.go`'s `Collect()` with a `slices.Contains()` guard. If the module should be enabled by default, also add it to the default modules list in `main.go` line 33

## Documentation Requirements
- **Every code change that adds, removes, or modifies metrics MUST also update the corresponding docs:**
  1. Update the metric table in `docs/metrics/` for the relevant module page (e.g. `docs/metrics/printer-objects.md`)
  2. If adding a new module, create a new page in `docs/metrics/` and register it in `docs/.vitepress/config.js` sidebar
  3. Update the summary table in `docs/metrics/index.md`
  4. If adding Moonraker API endpoint references, link to the official Moonraker docs
  5. Verify the site builds with `cd docs && npm run build`
  6. If adding, removing, or renaming metrics, also update the corresponding Grafana dashboard(s) in `test/grafana/provisioning/dashboards/` (panel queries, template variables, etc.)
- **New metric naming**: Follow `klipper_*_info` for labeled state gauges, `klipper_*` with `_total` suffix for counters, include units as suffixes (`_celsius`, `_mm`, `_seconds`)
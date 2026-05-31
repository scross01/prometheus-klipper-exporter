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
- `go run .` - Direct execution

## Architecture
- **Main Entry**: `main.go` handles HTTP server and routing
- **Collector Pattern**: `collector/collector.go` implements Prometheus Collector interface
- **Module System**: Each collector file (process_stats.go, job_queue.go, etc.) handles specific Klipper API endpoints
- **API Key Priority**: Header > CLI arg > Environment variable

## Non-Obvious Patterns
- **Metric Name Sanitization**: `GetValidLabelName()` in collector.go converts hyphens to underscores and strips invalid characters
- **Boolean Conversion**: `boolToFloat64()` converts booleans to 0/1 for Prometheus metrics
- **Module Registration**: Collector checks `slices.Contains(c.modules, "module_name")` to enable/disable features
- **API Response Handling**: Each collector file has `fetchMoonraker*` functions that handle HTTP requests and JSON parsing

## Code Style
- **Markdown**: Follows `.markdownlint.json` rules (setext_with_atx headers, line length)
- **Go**: Standard Go formatting with `go fmt`
- **Naming**: Prometheus metric names use `klipper_*` prefix with snake_case

## Critical Gotchas
- **Cross-Compilation**: Uses `GOOS`/`GOARCH` environment variables for platform-specific builds
- **Collector Pattern**: Must implement both `Describe()` and `Collect()` methods
- **Error Handling**: Returns early on errors but logs them first
- **Context Usage**: Collector uses context for cancellation and timeouts
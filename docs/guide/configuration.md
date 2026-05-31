# Configuration

## Command Line Options

### `-logging.level <level>`

Set the logging output verbosity. One of `Trace`, `Debug`, `Info`, `Warning`,
`Error`, `Fatal`, `Panic`. Default: `Info`

Can also be set with the `LOGGING_LEVEL` environment variable. The CLI flag
takes precedence.

### `-moonraker.apikey <string>`

API key for authenticating with Moonraker. See [Authentication](./authentication).

### `-web.listen-address [<ip>]:<port>`

Address to listen on for HTTP requests. Default: `:9101`

Examples:
- `:9101` — all interfaces, port 9101
- `192.168.1.99:7070` — specific IP and port

### `-help`

Display help text.

## Modules

Metric collection is organized into modules. Each module maps to a specific
Moonraker API endpoint. Modules are enabled via the `modules` parameter in the
Prometheus scrape configuration.

```yaml
params:
  modules:
    - process_stats
    - job_queue
    - system_info
    - network_stats
    - directory_info
    - printer_objects
    - history
    - mmu
```

If omitted, only the default modules are collected: `process_stats`, `job_queue`,
`system_info`.

| Module | Default | Description |
|--------|---------|-------------|
| `process_stats` | ✓ | Moonraker process and system CPU/memory metrics |
| `job_queue` | ✓ | Job queue length |
| `system_info` | ✓ | System CPU count |
| `network_stats` | | Network interface traffic and errors |
| `directory_info` | | Disk usage for gcodes directory |
| `history` | | Historical print job statistics |
| `printer_objects` | | Klipper printer object state (temperature, fans, MCU, etc.) |
| `mmu` | | Happy Hare Multi-Material Unit metrics |

The `temperature` module was deprecated in v0.8.0 and removed in v0.14.0 —
use `printer_objects` instead.

## Prometheus Scrape Configuration

See [Getting Started](./) for the full Prometheus configuration example.

### Multi-target exporter pattern

The exporter uses the Prometheus multi-target exporter pattern:

- `/metrics` — exporter's own metrics (process stats, Go runtime)
- `/probe?target=<klipper-host>:7125` — metrics for a specific Klipper instance

### API key in scrape config

Add the API key to the Prometheus scrape config using the `authorization` block:

```yaml
  - job_name: "klipper"
    authorization:
      type: APIKEY
      credentials: 'abcdef01234567890123456789012345'
      # credentials_file: /path/to/private/apikey.txt
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `LOGGING_LEVEL` | Log level (overridden by `-logging.level` flag) |
| `MOONRAKER_APIKEY` | Moonraker API key (lowest priority) |

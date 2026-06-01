# Network Stats

**Module:** `network_stats` (optional)  
**API Endpoint:** [`/machine/proc_stats`](https://moonraker.readthedocs.io/en/latest/web_api/#get-moonraker-process-stats) (network section)

Collects network interface statistics from the Klipper host. Each metric is
labelled with the network interface name.

## Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `klipper_network_rx_bytes` | Counter | `interface` | Received bytes |
| `klipper_network_tx_bytes` | Counter | `interface` | Transmitted bytes |
| `klipper_network_rx_packets` | Counter | `interface` | Received packets |
| `klipper_network_tx_packets` | Counter | `interface` | Transmitted packets |
| `klipper_network_rx_errs` | Counter | `interface` | Received packet errors |
| `klipper_network_tx_errs` | Counter | `interface` | Transmitted packet errors |
| `klipper_network_rx_drop` | Counter | `interface` | Received dropped packets |
| `klipper_network_tx_drop` | Counter | `interface` | Transmitted dropped packets |
| `klipper_network_bandwidth` | Gauge | `interface` | Current transmit bandwidth |

## Example PromQL

```promql
# Network throughput by interface (bits per second)
rate(klipper_network_rx_bytes{interface="wlan0"}[5m]) * 8

# Packet errors by interface
rate(klipper_network_rx_errs{interface="eth0"}[5m])
```

# Authentication

## Trusted Client

The simplest deployment option is to run the Klipper Exporter on a host that is
in the Moonraker trusted clients configuration. This is typically configured by
default to include all hosts on the local network.

If you have a more restrictive configuration, add the exporter host to the
`[authorization]` section of `moonraker.conf`:

```yaml
# moonraker.conf

[authorization]
trusted_clients:
  klipper-exporter
```

## API Key Authentication

Untrusted clients must use an API key to access Moonraker's HTTP APIs.

### Fetch the API key

Run the following on the Klipper host:

```sh
$ cd ~/moonraker/scripts
$ ./fetch-apikey.sh
abcdef01234567890123456789012345
```

### Provide the API key

The API key can be set in one of three ways, with the following priority order:

#### 1. Prometheus scrape configuration (highest priority)

```yaml
  - job_name: "klipper"
    authorization:
      type: APIKEY
      credentials: 'abcdef01234567890123456789012345'
```

> Only one API key can be set per job. For multiple Klipper hosts with different
> keys, create a separate job for each host.

#### 2. Command line argument

```sh
$ prometheus-klipper-exporter -moonraker.apikey='abcdef01234567890123456789012345'
```

#### 3. Environment variable (lowest priority)

```sh
$ export MOONRAKER_APIKEY='abcdef01234567890123456789012345'
$ prometheus-klipper-exporter
```

### Priority order

The exporter checks for the API key in this order:
1. `Authorization` header from the Prometheus scrape request (set via `authorization` in `prometheus.yml`)
2. `-moonraker.apikey` CLI flag
3. `MOONRAKER_APIKEY` environment variable

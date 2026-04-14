# go-healthchecks

A command line tool that runs health checks defined in a YAML config file and exits once they all pass. Designed for use in CI pipelines (e.g. GitLab CI `before_script`) to wait for services to become available before running tests or deployments.

## Usage

```bash
go-healthchecks check -c config.yaml
```

The process will retry checks every second until all pass (exit 0) or the timeout is reached (exit 1). The default timeout is 120 seconds.

## Config File

The config file defines one or more named health checks:

```yaml
---
checks:
  my-database:
    type: tcp
    config:
      host: localhost
      port: 5432
  my-api:
    type: http
    config:
      host: localhost
      port: 8443
      path: /healthz
      tls: true
  my-grpc-service:
    type: grpc
    config:
      host: localhost
      port: 50051
      certfile: /path/to/cert.pem
      keyfile: /path/to/key.pem
      cafile: /path/to/ca.pem
```

## Check Types

### tcp

Verifies a TCP connection can be established.

| Parameter | Description |
|-----------|-------------|
| `host`    | Hostname or IP |
| `port`    | Port number |

### http

Performs an HTTP GET request and considers any 2xx status code as healthy. Supports both HTTP and HTTPS via the `tls` flag, with optional mTLS client certificates.

| Parameter  | Description |
|------------|-------------|
| `host`     | Hostname or IP |
| `port`     | Port number |
| `path`     | URL path (e.g. `/healthz`) |
| `tls`      | Use HTTPS when `true`, HTTP when `false` |
| `certfile` | *(optional)* Path to client certificate for mTLS |
| `keyfile`  | *(optional)* Path to client private key for mTLS |
| `cafile`   | *(optional)* Path to CA certificate |

### grpc

Calls the standard gRPC health checking protocol (`grpc.health.v1.Health/Check`) over mTLS.

| Parameter  | Description |
|------------|-------------|
| `host`     | Hostname or IP |
| `port`     | Port number |
| `certfile` | Path to client certificate |
| `keyfile`  | Path to client private key |
| `cafile`   | Path to CA certificate |

## CI Example

GitLab CI `before_script` usage:

```yaml
before_script:
  - go-healthchecks check -c healthchecks.yaml
```

This blocks the job until all services are healthy, then continues with the rest of the script.

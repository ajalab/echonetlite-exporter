# AGENTS.md

## Project Overview

`echonetlite-exporter` is a Prometheus exporter for ECHONET Lite devices.
It discovers devices on the local network and exposes their metrics over HTTP
at `/metrics` (default listen address: `:9200`).

## Package Structure

### `main` (`main.go`)

Entry point of the application.

- Parses CLI flags (web listen address, multicast interface, discovery/collection timing)
- Starts HTTP server and Prometheus metrics endpoint
- Creates ECHONET Lite connection
- Discovers target devices using node profile scan
- Builds and starts collectors for each supported device class

### `internal/echonetlite`

ECHONET Lite protocol and device communication layer.

This package provides device clients/types for supported classes, such as PV power generation and power distribution board metering.

### `internal/collector`

Prometheus collection layer built on top of `internal/echonetlite`.

This package provides per-device-class collectors that poll ECHONET Lite properties at intervals.

## Notes for Agents

- Prefer extending `internal/echonetlite` first when adding support for a new
  ECHONET Lite class/property, then expose it via a collector in
  `internal/collector`.
- Register new collectors in `main.go` (`NewExporter`, `RegisterMetrics`,
  and `Start`).

# AGENTS.md

## Device support

- Implement ECHONET Lite class and property support in `internal/echonetlite`.
- Expose device metrics in `internal/collector`.
- Wire new device classes into discovery, collector registration, and startup in `main.go`.

## ECHONET Lite property reads

- Use ECHONET Lite property names for decoded Go struct fields.
- Treat requested EPCs as required unless documented as optional.
- Return an error for the entire read if a required property is missing or cannot be decoded.

## Prometheus metrics

- Prefix device metric names with `echonetlite_<device_class>_`.
- Use Prometheus base-unit suffixes such as `_watts` and `_joules`.
- Export instantaneous readings as gauges.
- Export cumulative readings as counters.
- End counter names with `_total`.
- Omit `instantaneous` and `cumulative` from metric names.
- Preserve the names, labels, units, and types of existing metrics.

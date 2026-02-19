# AGENTS.override.md for `internal/collector`

This file defines package-local conventions that take precedence for code under
`internal/collector`.

## Metric Naming Convention

Follow Prometheus naming best practices by default:

- Use lowercase snake_case metric names.
- Include base unit in the metric name where applicable (for example:
  `_watts`, `_joules`).
- Use `_total` suffix for counters.
- Keep label names consistent and stable (for example: `host`, `eoj`,
  optionally `channel`).

### Prefix

- All metrics in this package must start with `echonetlite_`.
- Device-specific metrics should then include a device namespace:
  `echonetlite_<device_class>_<property_and_unit>`.

Examples:

- `echonetlite_pv_power_generation_electric_power_generation_watts`
- `echonetlite_power_distribution_board_metering_electric_energy_simplex_joules_total`
- `echonetlite_storage_battery_ac_chargeable_electric_energy_joules`

## Instantaneous/Cumulative Term Rules

### Instantaneous properties

- Omit the word `instantaneous` from the metric name.
- Export as a `Gauge` (`prometheus.GaugeVec` or `prometheus.GaugeValue`).

Example:

- Property: `instantaneousElectricPowerGeneration`
- Metric: `..._electric_power_generation_watts` (Gauge)

### Cumulative properties

- Omit the word `cumulative` from the metric name.
- Export as a `Counter` (`prometheus.CounterValue` / counter-style metric).
- Metric name must end with `_total`.

Example:

- Property: `cumulativeElectricEnergyOfGeneration`
- Metric: `..._electric_energy_generation_joules_total` (Counter)

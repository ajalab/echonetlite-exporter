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

- Use `echonetlite_<device_object_class>_<device_object_property>` as the base for metric names.
- Use the full ECHONET Lite MRA `shortName` of the device object class in `snake_case` for `<device_object_class>`, as in `instantaneousWaterHeater` → `instantaneous_water_heater`.
- Use the ECHONET Lite MRA `shortName` of the device object property in `snake_case` for `<device_object_property>`.
- Omit `instantaneous` and `cumulative` only from `<device_object_property>` when they describe the reading.
- Append a Prometheus base-unit suffix to metrics with units, such as `_watts`, `_joules`, `_celsius`, or `_cubic_meters`.
- Omit the unit suffix for dimensionless states.
- Export instantaneous readings as gauges.
- Export cumulative readings as counters.
- End counter names with `_total` after any unit suffix, as in `echonetlite_watt_hour_meter_electric_energy_joules_total` for `wattHourMeter` / `cumulativeElectricEnergy`.
- For device metrics, use the ECHONET Lite MRA `propertyName.en` followed by a period and ` EPC=0xNN` as the Prometheus `HELP` text. Use two uppercase hexadecimal digits for the EPC.
- By default, export one time series per device object instance and property.
- Do not use MRA property values as metric label values.
- Preserve the names, labels, units, and types of existing metrics.

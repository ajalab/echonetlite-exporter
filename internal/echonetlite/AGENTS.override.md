# AGENTS.override.md for `internal/echonetlite`

This file defines package-local conventions that take precedence for code under
`internal/echonetlite`.

## Naming Convention

### Device client types

- Use `<DeviceClassName>Client` for ECHONET Lite client structs.
- Constructor name must be `New<DeviceClassName>Client`.
- Method for retrieval should be `Get(ctx context.Context, host string, eoj EOJ)`.

Examples:

- `PVPowerGenerationClient`, `NewPVPowerGenerationClient`
- `PowerDistributionBoardMeteringClient`, `NewPowerDistributionBoardMeteringClient`
- `StorageBatteryClient`, `NewStorageBatteryClient`

### Device properties/result types

- Use `<DeviceClassName>` for the parsed result struct returned by `Get`.
- Struct fields should use clear property names aligned to ECHONET Lite property semantics.
- Parse helper names should use `parse<PropertyName>` and parse exactly one property.

Examples:

- `PVPowerGeneration`
- `PowerDistributionBoardMetering`
- `StorageBattery`
- `parseInstantaneousElectricPowerGeneration`
- `parseCumulativeElectricEnergyOfGeneration`

## Fail-Fast Rule for Property Retrieval

For each device client `Get` call:

- Treat all requested EPC properties as required unless explicitly documented otherwise.
- If a required property is missing in the response, return an error.
- If parsing of any required property fails, return an error immediately.

# MRA Layout Reference

## Expected Directory Layout

This skill expects an ECHONET Lite MRA bundle that looks like:

- `devices/*.json`
- `superClass/0x0000.json`
- `nodeProfile/0x0EF0.json`
- `definitions/definitions.json`

Default path used by the wrapper script is `MRA_v1.3.1` relative to the current working directory.

## Object JSON Fields

Common top-level fields in object JSON files:

- `eoj`: EOJ string like `0x027D`
- `shortName`: object short name like `storageBattery`
- `className`: localized names, typically `ja` and `en`
- `elProperties`: array of property definitions

The wrapper searches object files across:

- `devices/`
- `superClass/`
- `nodeProfile/`

## Property JSON Fields

Each entry in `elProperties[]` commonly includes:

- `epc`: EPC string like `0xA3`
- `shortName`: property short name like `acDischargeableCapacity`
- `propertyName`: localized names (`ja`, `en`)
- `accessRule`: `get`, `set`, `inf`
- `validRelease`: release range (`from`, `to`)
- `data`: schema for EDT payload (inline schema or `$ref`)

## Data Reference Format

Property data often references `definitions.json`:

```json
{
  "data": {
    "$ref": "#/definitions/number_0-999999999Wh"
  }
}
```

Resolve with:

- File: `definitions/definitions.json`
- Path: `.definitions["number_0-999999999Wh"]`

## Observed Example

Storage battery object:

- `shortName`: `storageBattery`
- `eoj`: `0x027D`

Property example:

- `shortName`: `acDischargeableCapacity`
- `epc`: `0xA3`
- `data.$ref`: `#/definitions/number_0-999999999Wh`

Resolved definition example (from `definitions.json`):

- `type`: `number`
- `format`: `uint32`
- `minimum`: `0`
- `maximum`: `999999999`
- `unit`: `Wh`

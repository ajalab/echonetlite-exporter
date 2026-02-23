---
name: echonetlite-mra-query
description: Query the local ECHONET Lite Machine Readable Appendix (MRA) JSON dataset (`MRA_v1.3.1`) to list supported properties for a device (by device shortName or EOJ like `0x027D`) and fetch a specific property definition (by property shortName). Use when Codex needs to inspect ECHONET Lite class/property metadata or resolve `data.$ref` entries via `definitions/definitions.json`.
---

# ECHONET Lite MRA Query

## Overview

Use the bundled `scripts/mra_query.sh` wrapper to query the local MRA JSON files with `jq`.
Prefer the wrapper for routine lookups; use raw `jq` only for debugging or ad-hoc inspection.

## Quick Start

Run from the repository root (`/Users/koki/prj/echonetlite-exporter`) so the default `MRA_v1.3.1` path resolves automatically.

```bash
.codex/skills/echonetlite-mra-query/scripts/mra_query.sh list-props storageBattery
.codex/skills/echonetlite-mra-query/scripts/mra_query.sh list-props 0x027D --format table
.codex/skills/echonetlite-mra-query/scripts/mra_query.sh prop-def 0x027D acDischargeableCapacity
.codex/skills/echonetlite-mra-query/scripts/mra_query.sh prop-def 0x027D acDischargeableCapacity --resolve-data
```

If the current working directory is not the repo root, pass `--mra-dir /path/to/MRA_v1.3.1`.

## Device Lookup Rules

- Accept device shortName exactly as found in MRA JSON (example: `storageBattery`)
- Accept EOJ as `0xNNNN` (example: `0x027D`, case-insensitive)
- Search these object directories: `devices/`, `superClass/`, `nodeProfile/`
- Fail if zero matches or multiple matches are found

## Property Lookup Rules

- Property lookup for `prop-def` matches exact property `shortName` only
- Do not match EPC or localized property names unless the caller explicitly asks you to extend the script
- Use `list-props` first when you need to discover the property short name

## Resolve `data.$ref`

Use `--resolve-data` with `prop-def` to include a resolved schema from `definitions/definitions.json`.

- The script preserves the original property object
- It adds:
  - `dataRef`: the top-level `property.data.$ref` string (if present)
  - `resolvedData`: resolved schema for `property.data` with nested `#/definitions/...` references replaced recursively when encountered
- If `property.data` is inline and has no `$ref`, `resolvedData` is returned as `null`

## Examples

List storage battery properties:
```bash
.codex/skills/echonetlite-mra-query/scripts/mra_query.sh list-props storageBattery --format table
```

Fetch `acDischargeableCapacity` (`0xA3`) for storage battery by EOJ:
```bash
.codex/skills/echonetlite-mra-query/scripts/mra_query.sh prop-def 0x027D acDischargeableCapacity
```

Fetch the same property and resolve the referenced data definition:
```bash
.codex/skills/echonetlite-mra-query/scripts/mra_query.sh prop-def 0x027D acDischargeableCapacity --resolve-data
```

Debug with raw `jq` (manual inspection):
```bash
jq '.elProperties[] | select(.shortName=="acDischargeableCapacity")' MRA_v1.3.1/devices/0x027D.json
jq '.definitions["number_0-999999999Wh"]' MRA_v1.3.1/definitions/definitions.json
```

## Troubleshooting

- `jq: command not found`: install `jq` or use an environment where `jq` is available
- `MRA directory not found`: run from the repo root or pass `--mra-dir`
- `No device matched`: verify the device shortName/EOJ; check `MRA_v1.3.1/devices/*.json`
- `Property not found`: run `list-props <device>` and use the exact property `shortName`

## Resources

- `scripts/mra_query.sh`: primary wrapper for device/property lookups
- `references/mra-layout.md`: schema/layout notes for the local MRA bundle

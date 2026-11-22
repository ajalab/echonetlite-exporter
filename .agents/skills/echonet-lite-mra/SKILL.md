---
name: echonet-lite-mra
description: Look up ECHONET Lite device classes (EOJ), properties (EPC), access rules, and data formats in the Machine Readable Appendix (MRA).
---

The Machine Readable Appendix (MRA) is a JSON reference for the Detailed Requirements for ECHONET Device Objects.

Use `third_party/echonet-lite-mra/MRA_v1.4.0/` from the repository root. If the data is missing, run `./third_party/echonet-lite-mra/fetch.sh` from the repository root. See the [MRA README](../../../third_party/echonet-lite-mra/README.md) for download instructions, sources, and version details.

Look up device classes in `devices/`, common properties in `superClass/0x0000.json`, and shared data definitions in `definitions/definitions.json`.

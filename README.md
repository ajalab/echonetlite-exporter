# echonetlite-exporter

A Prometheus exporter that discovers [ECHONET Lite](https://echonet.jp/about/features/) devices and exposes their metrics for monitoring.

## Quickstart

### Run as container

Using `docker` command:

```bash
docker run \
  --name echonetlite-exporter \
  --network host \
  ghcr.io/ajalab/echonetlite-exporter:latest \
  -net.multicastInterface=eth0
```

Using Docker Compose:

```yaml
services:
  echonetlite-exporter:
    container_name: echonetlite-exporter
    image: ghcr.io/ajalab/echonetlite-exporter:latest
    network_mode: host
    command:
      - "-net.multicastInterface=eth0"
```

> [!IMPORTANT]
> When running as a normal process, the OS often picks the correct interface automatically.
> In a Docker container, multicast interface selection is less reliable, so explicitly set `-net.multicastInterface` to the host interface connected to your ECHONET Lite network.

### Run with `go install`

```bash
go install github.com/ajalab/echonetlite-exporter@latest
```

```bash
echonetlite-exporter
```

## Supported devices

### 住宅用太陽光発電 (Household solar power generation)

Prefix: `echonetlite_pv_power_generation`

|Metric family|ECHONET Lite property (shortName)|EPC|
|---|---|---|
|`<prefix>_electric_power_generation_watts`|instantaneousElectricPowerGeneration|0xE0|
|`<prefix>_electric_energy_generation_joules_total`|cumulativeElectricEnergyOfGeneration|0xE1|

### 分電盤メータリング (Power distribution board metering)

Prefix: `echonetlite_power_distribution_board_metering`

|Metric family|ECHONET Lite property (shortName)|EPC|
|---|---|---|
|`<prefix>_electric_energy_simplex_joules_total`|cumulativeElectricEnergyListSimplex|0xB3|
|`<prefix>_electric_power_simplex_watts`|instantaneousElectricPowerListSimplex|0xB7|

### 分散型電源電力量メータ (Distributed generator's electric energy meter)

Prefix: `echonetlite_dr_electric_energy_meter`

|Metric family|ECHONET Lite property (shortName)|EPC|
|---|---|---|
|`<prefix>_ac_input_electric_energy_joules_total`|acInputCumulativeElectricEnergy|0xE0|
|`<prefix>_ac_output_electric_energy_joules_total`|acOutputCumulativeElectricEnergy|0xE2|
|`<prefix>_independent_operation_electric_energy_joules_total`|independentOperationCumulativeElectricEnergy|0xE4|
|`<prefix>_ac_inout_electric_power_watts`|acInstantaneousElectricPower|0xE9|
|`<prefix>_independent_operation_electric_power_watts`|independentOperationInstantaneousElectricPower|0xEA|

Notes:
- Cumulative energy values are scaled by `0xD4` (`cumulativeAmountsOfElectricEnergyUnit`) and exported as joules.
- This exporter treats missing requested DR meter properties and `NoData` values (`E4`, `E9`, `EA`) as collection failures.

### 蓄電池 (Storage battery)

Prefix: `echonetlite_storage_battery`

|Metric family|ECHONET Lite property (shortName)|EPC|
|---|---|---|
|`<prefix>_ac_chargeable_electric_energy_joules`|acChargeableElectricEnergy|0xA4|
|`<prefix>_ac_dischargeable_electric_energy_joules`|acDischargeableElectricEnergy|0xA5|
|`<prefix>_ac_charging_electric_energy_joules_total`|acCumulativeChargingElectricEnergy|0xA8|
|`<prefix>_ac_discharging_electric_energy_joules_total`|acCumulativeDischargingElectricEnergy|0xA9|

### マルチ入力PCS (Multiple input PCS)

Prefix: `echonetlite_multiple_input_pcs`

|Metric family|ECHONET Lite property (shortName)|EPC|
|---|---|---|
|`<prefix>_normal_direction_electric_energy_joules_total`|normalDirectionElectricEnergy|0xE0|
|`<prefix>_reverse_direction_electric_energy_joules_total`|reverseDirectionElectricEnergy|0xE3|
|`<prefix>_electric_power_watts`|instantaneousElectricPower|0xE7|

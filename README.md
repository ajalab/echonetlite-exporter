# echonetlite-exporter

A Prometheus exporter that discovers [ECHONET Lite](https://echonet.jp/about/features/) devices and exposes their metrics for monitoring.

## Quickstart

### Run with Docker

Start the exporter with the Docker CLI.

```bash
docker run \
  --name echonetlite-exporter \
  --network host \
  ghcr.io/ajalab/echonetlite-exporter:latest \
  -net.multicastInterface=eth0
```

Start the exporter with Docker Compose.

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
echonetlite-exporter
```

## Supported devices

### 住宅用太陽光発電 (Household solar power generation, `0x0279`)

Metric names use the `echonetlite_pv_power_generation` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_electric_power_generation_watts` | `instantaneousElectricPowerGeneration` | `0xE0` |
| `<prefix>_electric_energy_of_generation_joules_total` | `cumulativeElectricEnergyOfGeneration` | `0xE1` |

### 分電盤メータリング (Power distribution board metering, `0x0287`)

Metric names use the `echonetlite_power_distribution_board_metering` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_electric_energy_list_simplex_joules_total` | `cumulativeElectricEnergyListSimplex` | `0xB3` |
| `<prefix>_electric_power_list_simplex_watts` | `instantaneousElectricPowerListSimplex` | `0xB7` |

### 分散型電源電力量メータ (Distributed generator's electric energy meter, `0x028E`)

Metric names use the `echonetlite_dr_electric_energy_meter` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_ac_input_electric_energy_joules_total` | `acInputCumulativeElectricEnergy` | `0xE0` |
| `<prefix>_ac_output_electric_energy_joules_total` | `acOutputCumulativeElectricEnergy` | `0xE2` |
| `<prefix>_independent_operation_electric_energy_joules_total` | `independentOperationCumulativeElectricEnergy` | `0xE4` |
| `<prefix>_ac_electric_power_watts` | `acInstantaneousElectricPower` | `0xE9` |
| `<prefix>_independent_operation_electric_power_watts` | `independentOperationInstantaneousElectricPower` | `0xEA` |

### 蓄電池 (Storage battery, `0x027D`)

Metric names use the `echonetlite_storage_battery` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_ac_chargeable_electric_energy_joules` | `acChargeableElectricEnergy` | `0xA4` |
| `<prefix>_ac_dischargeable_electric_energy_joules` | `acDischargeableElectricEnergy` | `0xA5` |
| `<prefix>_ac_charging_electric_energy_joules_total` | `acCumulativeChargingElectricEnergy` | `0xA8` |
| `<prefix>_ac_discharging_electric_energy_joules_total` | `acCumulativeDischargingElectricEnergy` | `0xA9` |

### マルチ入力PCS (Multiple input PCS, `0x02A5`)

Metric names use the `echonetlite_multiple_input_pcs` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_normal_direction_electric_energy_joules_total` | `normalDirectionElectricEnergy` | `0xE0` |
| `<prefix>_reverse_direction_electric_energy_joules_total` | `reverseDirectionElectricEnergy` | `0xE3` |
| `<prefix>_electric_power_watts` | `instantaneousElectricPower` | `0xE7` |

### 瞬間式給湯器 (Instantaneous water heater, `0x0272`)

Metric names use the `echonetlite_instantaneous_water_heater` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_hot_water_heating_status` | `hotWaterHeatingStatus` | `0xD0` |
| `<prefix>_bath_water_heating_status` | `bathWaterHeatingStatus` | `0xE2` |
| `<prefix>_automatic_bath_operation` | `automaticBathOperation` | `0xE3` |
| `<prefix>_target_supplied_water_temperature_celsius` | `targetSuppliedWaterTemperature` | `0xD1` |
| `<prefix>_target_bath_water_temperature_celsius` | `targetBathWaterTemperature` | `0xE1` |
| `<prefix>_shower_hot_water_supply_status` | `showerHotWaterSupplyStatus` | `0xEA` |
| `<prefix>_kitchen_hot_water_supply_status` | `kitchenHotWaterSupplyStatus` | `0xEB` |
| `<prefix>_bath_operation_status` | `bathOperationStatus` | `0xEF` |

### 電力量メータ (Watt-hour meter, `0x0280`)

Metric names use the `echonetlite_watt_hour_meter` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_electric_energy_joules_total` | `cumulativeElectricEnergy` | `0xE0` |

### 水流量メータ (Water flowmeter, `0x0281`)

Metric names use the `echonetlite_water_flow_meter` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_flowing_water_cubic_meters_total` | `cumulativeFlowingWater` | `0xE0` |

### ガスメータ (Gas meter, `0x0282`)

Metric names use the `echonetlite_gas_meter` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_gas_consumption_cubic_meters_total` | `cumulativeGasConsumption` | `0xE0` |

### ハイブリッド給湯機 (Hybrid water heater, `0x02A6`)

Metric names use the `echonetlite_hybrid_water_heater` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_linkage_mode_for_solar_power_generation` | `linkageModeForSolarPowerGeneration` | `0xB8` |
| `<prefix>_water_heating_status` | `waterHeatingStatus` | `0xB2` |
| `<prefix>_heater_status` | `heaterStatus` | `0xB3` |
| `<prefix>_hot_water_supply_status` | `hotWaterSupplyStatus` | `0xC3` |
| `<prefix>_remaining_water_cubic_meters` | `remainingWater` | `0xE1` |
| `<prefix>_tank_capacity_cubic_meters` | `tankCapacity` | `0xE2` |

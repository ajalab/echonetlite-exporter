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

### Common device properties

These families use the prefix of each supported class.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_operation_status` | `operationStatus` | `0x80` |
| `<prefix>_fault_status` | `faultStatus` | `0x88` |

### 住宅用太陽光発電 (Household solar power generation, `0x0279`)

Metric names use the `echonetlite_pv_power_generation` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_fault_description` | `faultDescription` | `0x89` |
| `<prefix>_output_power_control1_ratio` | `outputPowerControl1` | `0xA0` |
| `<prefix>_output_power_control2_watts` | `outputPowerControl2` | `0xA1` |
| `<prefix>_upper_limit_clip_watts` | `upperLimitClip` | `0xB4` |
| `<prefix>_approved_capacity_watts` | `approvedCapacity` | `0xC3` |
| `<prefix>_power_system_interconnection_status` | `powerSystemInterconnectionStatus` | `0xD0` |
| `<prefix>_output_power_restraint_status` | `outputPowerRestraintStatus` | `0xD1` |
| `<prefix>_electric_power_generation_watts` | `instantaneousElectricPowerGeneration` | `0xE0` |
| `<prefix>_electric_energy_of_generation_joules_total` | `cumulativeElectricEnergyOfGeneration` | `0xE1` |
| `<prefix>_electric_energy_sold_joules_total` | `cumulativeElectricEnergySold` | `0xE3` |
| `<prefix>_power_generation_output_limit1_ratio` | `powerGenerationOutputLimit1` | `0xE5` |
| `<prefix>_power_generation_output_limit2_watts` | `powerGenerationOutputLimit2` | `0xE6` |
| `<prefix>_limit_electric_energy_sold_watts` | `limitElectricEnergySold` | `0xE7` |
| `<prefix>_rated_electric_power_ofgeneration_watts` | `ratedElectricPowerOfgeneration` | `0xE8` |
| `<prefix>_rated_electric_power_ofgeneration_independent_watts` | `ratedElectricPowerOfgenerationIndependent` | `0xE9` |

### 分電盤メータリング (Power distribution board metering, `0x0287`)

Metric names use the `echonetlite_power_distribution_board_metering` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_master_rated_capacity_amperes` | `masterRatedCapacity` | `0xB0` |
| `<prefix>_number_of_measurement_channels_simplex` | `numberOfMeasurementChannelsSimplex` | `0xB1` |
| `<prefix>_electric_energy_list_simplex_joules_total` | `cumulativeElectricEnergyListSimplex` | `0xB3` |
| `<prefix>_current_list_simplex_amperes` | `instantaneousCurrentListSimplex` | `0xB5` |
| `<prefix>_electric_power_list_simplex_watts` | `instantaneousElectricPowerListSimplex` | `0xB7` |
| `<prefix>_number_of_measurement_channels_duplex` | `numberOfMeasurementChannelsDuplex` | `0xB8` |
| `<prefix>_electric_energy_list_duplex_joules_total` | `cumulativeElectricEnergyListDuplex` | `0xBA` |
| `<prefix>_current_list_duplex_amperes` | `instantaneousCurrentListDuplex` | `0xBC` |
| `<prefix>_electric_power_list_duplex_watts` | `instantaneousElectricPowerListDuplex` | `0xBE` |
| `<prefix>_normal_direction_electric_energy_joules_total` | `normalDirectionCumulativeElectricEnergy` | `0xC0` |
| `<prefix>_reverse_direction_electric_energy_joules_total` | `reverseDirectionCumulativeElectricEnergy` | `0xC1` |
| `<prefix>_electric_power_watts` | `instantaneousElectricPower` | `0xC6` |
| `<prefix>_current_amperes` | `instantaneousCurrent` | `0xC7` |
| `<prefix>_voltage_volts` | `instantaneousVoltage` | `0xC8` |
| `<prefix>_measurement_channel<n>_amperes` | `measurementChannel1` – `measurementChannel32` | `0xD0` – `0xEF` |
| `<prefix>_measurement_channel<n>_joules_total` | `measurementChannel1` – `measurementChannel32` | `0xD0` – `0xEF` |

### 分散型電源電力量メータ (Distributed generator's electric energy meter, `0x028E`)

Metric names use the `echonetlite_dr_electric_energy_meter` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_time_synchronization_status` | `timeSynchronizationStatus` | `0xDB` |
| `<prefix>_ac_input_electric_energy_joules_total` | `acInputCumulativeElectricEnergy` | `0xE0` |
| `<prefix>_ac_output_electric_energy_joules_total` | `acOutputCumulativeElectricEnergy` | `0xE2` |
| `<prefix>_independent_operation_electric_energy_joules_total` | `independentOperationCumulativeElectricEnergy` | `0xE4` |
| `<prefix>_ac_electric_power_watts` | `acInstantaneousElectricPower` | `0xE9` |
| `<prefix>_independent_operation_electric_power_watts` | `independentOperationInstantaneousElectricPower` | `0xEA` |

### 蓄電池 (Storage battery, `0x027D`)

Metric names use the `echonetlite_storage_battery` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_fault_description` | `faultDescription` | `0x89` |
| `<prefix>_ac_effective_charging_capacity_joules` | `acEffectiveChargingCapacity` | `0xA0` |
| `<prefix>_ac_effective_discharging_capacity_joules` | `acEffectiveDischargingCapacity` | `0xA1` |
| `<prefix>_ac_chargeable_capacity_joules` | `acChargeableCapacity` | `0xA2` |
| `<prefix>_ac_dischargeable_capacity_joules` | `acDischargeableCapacity` | `0xA3` |
| `<prefix>_ac_chargeable_electric_energy_joules` | `acChargeableElectricEnergy` | `0xA4` |
| `<prefix>_ac_dischargeable_electric_energy_joules` | `acDischargeableElectricEnergy` | `0xA5` |
| `<prefix>_ac_charge_upper_limit_ratio` | `acChargeUpperLimit` | `0xA6` |
| `<prefix>_ac_discharge_lower_limit_ratio` | `acDischargeLowerLimit` | `0xA7` |
| `<prefix>_ac_charging_electric_energy_joules_total` | `acCumulativeChargingElectricEnergy` | `0xA8` |
| `<prefix>_ac_discharging_electric_energy_joules_total` | `acCumulativeDischargingElectricEnergy` | `0xA9` |
| `<prefix>_ac_target_charging_electric_energy_joules` | `acTargetChargingElectricEnergy` | `0xAA` |
| `<prefix>_ac_target_discharging_electric_energy_joules` | `acTargetDischargingElectricEnergy` | `0xAB` |
| `<prefix>_charging_method` | `chargingMethod` | `0xC1` |
| `<prefix>_discharging_method` | `dischargingMethod` | `0xC2` |
| `<prefix>_ac_rated_electric_energy_joules` | `acRatedElectricEnergy` | `0xC7` |
| `<prefix>_minimum_and_maximum_charging_electric_power_watts` | `minimumAndMaximumChargingElectricPower` | `0xC8` |
| `<prefix>_minimum_and_maximum_discharging_electric_power_watts` | `minimumAndMaximumDischargingElectricPower` | `0xC9` |
| `<prefix>_minimum_and_maximum_charging_current_amperes` | `minimumAndMaximumChargingCurrent` | `0xCA` |
| `<prefix>_minimum_and_maximum_discharging_current_amperes` | `minimumAndMaximumDischargingCurrent` | `0xCB` |
| `<prefix>_re_interconnection_permission` | `reInterconnectionPermission` | `0xCC` |
| `<prefix>_operation_permission` | `operationPermission` | `0xCD` |
| `<prefix>_independent_operation_permission` | `independentOperationPermission` | `0xCE` |
| `<prefix>_actual_operation_mode` | `actualOperationMode` | `0xCF` |
| `<prefix>_rated_electric_energy_joules` | `ratedElectricEnergy` | `0xD0` |
| `<prefix>_rated_capacity_coulombs` | `ratedCapacity` | `0xD1` |
| `<prefix>_rated_voltage_volts` | `ratedVoltage` | `0xD2` |
| `<prefix>_charging_and_discharging_electric_power_watts` | `instantaneousChargingAndDischargingElectricPower` | `0xD3` |
| `<prefix>_charging_and_discharging_current_amperes` | `instantaneousChargingAndDischargingCurrent` | `0xD4` |
| `<prefix>_charging_and_discharging_voltage_volts` | `instantaneousChargingAndDischargingVoltage` | `0xD5` |
| `<prefix>_discharging_electric_energy_joules_total` | `cumulativeDischargingElectricEnergy` | `0xD6` |
| `<prefix>_charging_electric_energy_joules_total` | `cumulativeChargingElectricEnergy` | `0xD8` |
| `<prefix>_operation_mode` | `operationMode` | `0xDA` |
| `<prefix>_power_system_interconnection_status` | `powerSystemInterconnectionStatus` | `0xDB` |
| `<prefix>_minimum_and_maximum_charging_power_at_independent_watts` | `minimumAndMaximumChargingPowerAtIndependent` | `0xDC` |
| `<prefix>_minimum_and_maximum_discharging_power_at_independent_watts` | `minimumAndMaximumDischargingPowerAtIndependent` | `0xDD` |
| `<prefix>_minimum_and_maximum_charging_current_at_independent_amperes` | `minimumAndMaximumChargingCurrentAtIndependent` | `0xDE` |
| `<prefix>_minimum_and_maximum_discharging_current_at_independent_amperes` | `minimumAndMaximumDischargingCurrentAtIndependent` | `0xDF` |
| `<prefix>_remaining_capacity1_joules` | `remainingCapacity1` | `0xE2` |
| `<prefix>_remaining_capacity2_coulombs` | `remainingCapacity2` | `0xE3` |
| `<prefix>_remaining_capacity3_ratio` | `remainingCapacity3` | `0xE4` |
| `<prefix>_battery_health_state_ratio` | `batteryHealthState` | `0xE5` |

### マルチ入力PCS (Multiple input PCS, `0x02A5`)

Metric names use the `echonetlite_multiple_input_pcs` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_fault_description` | `faultDescription` | `0x89` |
| `<prefix>_power_system_interconnection_status` | `powerSystemInterconnectionStatus` | `0xD0` |
| `<prefix>_normal_direction_electric_energy_joules_total` | `normalDirectionElectricEnergy` | `0xE0` |
| `<prefix>_reverse_direction_electric_energy_joules_total` | `reverseDirectionElectricEnergy` | `0xE3` |
| `<prefix>_electric_power_watts` | `instantaneousElectricPower` | `0xE7` |

### 瞬間式給湯器 (Instantaneous water heater, `0x0272`)

Metric names use the `echonetlite_instantaneous_water_heater` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_hot_water_heating_status` | `hotWaterHeatingStatus` | `0xD0` |
| `<prefix>_target_supplied_water_temperature_celsius` | `targetSuppliedWaterTemperature` | `0xD1` |
| `<prefix>_hot_water_warmer_setting` | `hotWaterWarmerSetting` | `0xD2` |
| `<prefix>_remaining_automatic_operation_time_seconds` | `remainingAutomaticOperationTime` | `0xDB` |
| `<prefix>_target_bath_water_temperature_celsius` | `targetBathWaterTemperature` | `0xE1` |
| `<prefix>_bath_water_heating_status` | `bathWaterHeatingStatus` | `0xE2` |
| `<prefix>_automatic_bath_operation` | `automaticBathOperation` | `0xE3` |
| `<prefix>_target_bath_additional_boilup_operation` | `targetBathAdditionalBoilupOperation` | `0xE4` |
| `<prefix>_bath_hot_water_addition` | `bathHotWaterAddition` | `0xE5` |
| `<prefix>_bath_lukewarm_water_function` | `bathLukewarmWaterFunction` | `0xE6` |
| `<prefix>_bath_water_volume1_cubic_meters` | `bathWaterVolume1` | `0xE7` |
| `<prefix>_bathroom_priority` | `bathroomPriority` | `0xE9` |
| `<prefix>_shower_hot_water_supply_status` | `showerHotWaterSupplyStatus` | `0xEA` |
| `<prefix>_kitchen_hot_water_supply_status` | `kitchenHotWaterSupplyStatus` | `0xEB` |
| `<prefix>_bath_water_volume3_cubic_meters` | `bathWaterVolume3` | `0xEE` |
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
| `<prefix>_abnormal_detection` | `abnormalDetection` | `0xE3` |
| `<prefix>_security_data_information` | `securityDataInformation` | `0xE4` |

### ガスメータ (Gas meter, `0x0282`)

Metric names use the `echonetlite_gas_meter` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_gas_consumption_cubic_meters_total` | `cumulativeGasConsumption` | `0xE0` |

### ハイブリッド給湯機 (Hybrid water heater, `0x02A6`)

Metric names use the `echonetlite_hybrid_water_heater` prefix.

| Metric family | ECHONET Lite property (`shortName`) | EPC |
| --- | --- | --- |
| `<prefix>_automatic_water_heating` | `automaticWaterHeating` | `0xB0` |
| `<prefix>_water_heating_status` | `waterHeatingStatus` | `0xB2` |
| `<prefix>_heater_status` | `heaterStatus` | `0xB3` |
| `<prefix>_hot_water_supply_mode_for_auxiliary_heat_source_machine` | `hotWaterSupplyModeForAuxiliaryHeatSourceMachine` | `0xB6` |
| `<prefix>_heater_mode_for_auxiliary_heat_source_machine` | `heaterModeForAuxiliaryHeatSourceMachine` | `0xB7` |
| `<prefix>_linkage_mode_for_solar_power_generation` | `linkageModeForSolarPowerGeneration` | `0xB8` |
| `<prefix>_hot_water_supply_status` | `hotWaterSupplyStatus` | `0xC3` |
| `<prefix>_remaining_water_cubic_meters` | `remainingWater` | `0xE1` |
| `<prefix>_tank_capacity_cubic_meters` | `tankCapacity` | `0xE2` |

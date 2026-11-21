package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestInstantaneousWaterHeaterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.InstantaneousWaterHeater{
		HotWaterHeatingStatus:          true,
		AutomaticBathOperation:         new(false),
		TargetSuppliedWaterTemperature: new(uint8(42)),
		TargetBathWaterTemperature:     new(uint8(40)),
		ShowerHotWaterSupplyStatus:     new(true),
		KitchenHotWaterSupplyStatus:    new(false),
		BathOperationStatus:            new(uint8(0x43)),
	}
	const want = `
# HELP echonetlite_instantaneous_water_heater_automatic_bath_operation Bath Auto mode setting.
# TYPE echonetlite_instantaneous_water_heater_automatic_bath_operation gauge
echonetlite_instantaneous_water_heater_automatic_bath_operation{eoj="027201",host="192.0.2.10"} 0
# HELP echonetlite_instantaneous_water_heater_bath_operation_status Bath operation status monitor.
# TYPE echonetlite_instantaneous_water_heater_bath_operation_status gauge
echonetlite_instantaneous_water_heater_bath_operation_status{eoj="027201",host="192.0.2.10"} 67
# HELP echonetlite_instantaneous_water_heater_bath_water_heating_status Bath water heater status.
# TYPE echonetlite_instantaneous_water_heater_bath_water_heating_status gauge
echonetlite_instantaneous_water_heater_bath_water_heating_status{eoj="027201",host="192.0.2.10"} 0
# HELP echonetlite_instantaneous_water_heater_hot_water_heating_status Hot water heating status.
# TYPE echonetlite_instantaneous_water_heater_hot_water_heating_status gauge
echonetlite_instantaneous_water_heater_hot_water_heating_status{eoj="027201",host="192.0.2.10"} 1
# HELP echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status Kitchen hot water supply status.
# TYPE echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status gauge
echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status{eoj="027201",host="192.0.2.10"} 0
# HELP echonetlite_instantaneous_water_heater_shower_hot_water_supply_status Shower hot water supply status.
# TYPE echonetlite_instantaneous_water_heater_shower_hot_water_supply_status gauge
echonetlite_instantaneous_water_heater_shower_hot_water_supply_status{eoj="027201",host="192.0.2.10"} 1
# HELP echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius Set value of bath temperature.
# TYPE echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius gauge
echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius{eoj="027201",host="192.0.2.10"} 40
# HELP echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius Set value of hot water temperature.
# TYPE echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius gauge
echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius{eoj="027201",host="192.0.2.10"} 42
`
	assertPublishedMetrics(t, NewInstantaneousWaterHeaterCollector, echonetlite.ClassInstantaneousWaterHeater, reading, want)
}

package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestInstantaneousWaterHeaterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.InstantaneousWaterHeater{
		SuperClass:                     echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		HotWaterHeatingStatus:          0x41,
		BathWaterHeatingStatus:         0x42,
		AutomaticBathOperation:         new(uint8(0x42)),
		TargetSuppliedWaterTemperature: new(uint8(42)),
		TargetBathWaterTemperature:     new(uint8(40)),
		ShowerHotWaterSupplyStatus:     new(uint8(0x41)),
		KitchenHotWaterSupplyStatus:    new(uint8(0x42)),
		BathOperationStatus:            new(uint8(0x43)),
	}
	const want = `
# HELP echonetlite_instantaneous_water_heater_automatic_bath_operation Bath Auto mode setting. values={"65":"Auto ON","66":"Auto OFF"} EPC=0xE3
# TYPE echonetlite_instantaneous_water_heater_automatic_bath_operation gauge
echonetlite_instantaneous_water_heater_automatic_bath_operation{eoj="027201",host="192.0.2.10"} 66
# HELP echonetlite_instantaneous_water_heater_bath_operation_status Bath operation status monitor. values={"65":"Supplying hot water","66":"Stopped","67":"Keeping bath temperature"} EPC=0xEF
# TYPE echonetlite_instantaneous_water_heater_bath_operation_status gauge
echonetlite_instantaneous_water_heater_bath_operation_status{eoj="027201",host="192.0.2.10"} 67
# HELP echonetlite_instantaneous_water_heater_bath_water_heating_status Bath water heater status. values={"65":"Heating","66":"Not Heating"} EPC=0xE2
# TYPE echonetlite_instantaneous_water_heater_bath_water_heating_status gauge
echonetlite_instantaneous_water_heater_bath_water_heating_status{eoj="027201",host="192.0.2.10"} 66
# HELP echonetlite_instantaneous_water_heater_hot_water_heating_status Hot water heating status. values={"65":"Hot water heating status found","66":"Hot water heating status not found"} EPC=0xD0
# TYPE echonetlite_instantaneous_water_heater_hot_water_heating_status gauge
echonetlite_instantaneous_water_heater_hot_water_heating_status{eoj="027201",host="192.0.2.10"} 65
# HELP echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status Kitchen hot water supply status. values={"65":"Kitchen hot water supply ON","66":"Kitchen hot water supply OFF"} EPC=0xEB
# TYPE echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status gauge
echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status{eoj="027201",host="192.0.2.10"} 66
# HELP echonetlite_instantaneous_water_heater_shower_hot_water_supply_status Shower hot water supply status. values={"65":"Shower hot water supply ON","66":"Shower hot water supply OFF"} EPC=0xEA
# TYPE echonetlite_instantaneous_water_heater_shower_hot_water_supply_status gauge
echonetlite_instantaneous_water_heater_shower_hot_water_supply_status{eoj="027201",host="192.0.2.10"} 65
# HELP echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius Set value of bath temperature. EPC=0xE1
# TYPE echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius gauge
echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius{eoj="027201",host="192.0.2.10"} 40
# HELP echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius Set value of hot water temperature. EPC=0xD1
# TYPE echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius gauge
echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius{eoj="027201",host="192.0.2.10"} 42
# HELP echonetlite_instantaneous_water_heater_fault_status Fault status. values={"65":"Fault occurred.","66":"No fault has occurred."} EPC=0x88
# TYPE echonetlite_instantaneous_water_heater_fault_status gauge
echonetlite_instantaneous_water_heater_fault_status{eoj="027201",host="192.0.2.10"} 66
# HELP echonetlite_instantaneous_water_heater_operation_status Operation status. values={"48":"ON","49":"OFF"} EPC=0x80
# TYPE echonetlite_instantaneous_water_heater_operation_status gauge
echonetlite_instantaneous_water_heater_operation_status{eoj="027201",host="192.0.2.10"} 48
`
	assertPublishedMetrics(t, NewInstantaneousWaterHeaterCollector, echonetlite.ClassInstantaneousWaterHeater, reading, want)
}

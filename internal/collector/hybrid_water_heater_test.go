package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestHybridWaterHeaterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.HybridWaterHeater{
		SuperClass:                         echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		LinkageModeForSolarPowerGeneration: 0x42,
		WaterHeatingStatus:                 new(uint8(0x41)), HeaterStatus: new(uint8(0x42)), HotWaterSupplyStatus: new(uint8(0x41)),
		RemainingWater: new(uint16(150)), TankCapacity: new(uint16(200)),
	}
	const want = `
# HELP echonetlite_hybrid_water_heater_heater_status Heater status. values={"65":"Water is heating.","66":"Water is not heating."} EPC=0xB3
# TYPE echonetlite_hybrid_water_heater_heater_status gauge
echonetlite_hybrid_water_heater_heater_status{eoj="02A601",host="192.0.2.10"} 66
# HELP echonetlite_hybrid_water_heater_hot_water_supply_status Hot water supply status. values={"65":"Supplying hot water","66":"Not supplying hot water"} EPC=0xC3
# TYPE echonetlite_hybrid_water_heater_hot_water_supply_status gauge
echonetlite_hybrid_water_heater_hot_water_supply_status{eoj="02A601",host="192.0.2.10"} 65
# HELP echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation Linkage mode setting for solar power generation. values={"65":"Mode off","66":"Household consumption","67":"Prioritizing electricity sales","68":"Economic efficiency"} EPC=0xB8
# TYPE echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation gauge
echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation{eoj="02A601",host="192.0.2.10"} 66
# HELP echonetlite_hybrid_water_heater_remaining_water_cubic_meters Measured amount of hot water remaining in tank. EPC=0xE1
# TYPE echonetlite_hybrid_water_heater_remaining_water_cubic_meters gauge
echonetlite_hybrid_water_heater_remaining_water_cubic_meters{eoj="02A601",host="192.0.2.10"} 0.15
# HELP echonetlite_hybrid_water_heater_tank_capacity_cubic_meters Tank capacity. EPC=0xE2
# TYPE echonetlite_hybrid_water_heater_tank_capacity_cubic_meters gauge
echonetlite_hybrid_water_heater_tank_capacity_cubic_meters{eoj="02A601",host="192.0.2.10"} 0.2
# HELP echonetlite_hybrid_water_heater_water_heating_status Water heating status. values={"65":"Water is heating.","66":"Water is not heating."} EPC=0xB2
# TYPE echonetlite_hybrid_water_heater_water_heating_status gauge
echonetlite_hybrid_water_heater_water_heating_status{eoj="02A601",host="192.0.2.10"} 65
# HELP echonetlite_hybrid_water_heater_fault_status Fault status. values={"65":"Fault occurred.","66":"No fault has occurred."} EPC=0x88
# TYPE echonetlite_hybrid_water_heater_fault_status gauge
echonetlite_hybrid_water_heater_fault_status{eoj="02A601",host="192.0.2.10"} 66
# HELP echonetlite_hybrid_water_heater_operation_status Operation status. values={"48":"ON","49":"OFF"} EPC=0x80
# TYPE echonetlite_hybrid_water_heater_operation_status gauge
echonetlite_hybrid_water_heater_operation_status{eoj="02A601",host="192.0.2.10"} 48
`
	assertPublishedMetrics(t, NewHybridWaterHeaterCollector, echonetlite.ClassHybridWaterHeater, reading, want)
}

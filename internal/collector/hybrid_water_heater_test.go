package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestHybridWaterHeaterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.HybridWaterHeater{
		LinkageModeForSolarPowerGeneration: 0x42,
		WaterHeatingStatus:                 new(true), HeaterStatus: new(false), HotWaterSupplyStatus: new(true),
		RemainingWater: new(uint16(150)), TankCapacity: new(uint16(200)),
	}
	const want = `
# HELP echonetlite_hybrid_water_heater_heater_status Heater status. EPC=0xB3
# TYPE echonetlite_hybrid_water_heater_heater_status gauge
echonetlite_hybrid_water_heater_heater_status{eoj="02A601",host="192.0.2.10"} 0
# HELP echonetlite_hybrid_water_heater_hot_water_supply_status Hot water supply status. EPC=0xC3
# TYPE echonetlite_hybrid_water_heater_hot_water_supply_status gauge
echonetlite_hybrid_water_heater_hot_water_supply_status{eoj="02A601",host="192.0.2.10"} 1
# HELP echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation Linkage mode setting for solar power generation. EPC=0xB8
# TYPE echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation gauge
echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation{eoj="02A601",host="192.0.2.10"} 66
# HELP echonetlite_hybrid_water_heater_remaining_water_cubic_meters Measured amount of hot water remaining in tank. EPC=0xE1
# TYPE echonetlite_hybrid_water_heater_remaining_water_cubic_meters gauge
echonetlite_hybrid_water_heater_remaining_water_cubic_meters{eoj="02A601",host="192.0.2.10"} 0.15
# HELP echonetlite_hybrid_water_heater_tank_capacity_cubic_meters Tank capacity. EPC=0xE2
# TYPE echonetlite_hybrid_water_heater_tank_capacity_cubic_meters gauge
echonetlite_hybrid_water_heater_tank_capacity_cubic_meters{eoj="02A601",host="192.0.2.10"} 0.2
# HELP echonetlite_hybrid_water_heater_water_heating_status Water heating status. EPC=0xB2
# TYPE echonetlite_hybrid_water_heater_water_heating_status gauge
echonetlite_hybrid_water_heater_water_heating_status{eoj="02A601",host="192.0.2.10"} 1
`
	assertPublishedMetrics(t, NewHybridWaterHeaterCollector, echonetlite.ClassHybridWaterHeater, reading, want)
}

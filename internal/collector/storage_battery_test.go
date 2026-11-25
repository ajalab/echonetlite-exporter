package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestStorageBatteryCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.StorageBattery{
		SuperClass:                             echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		ActualOperationMode:                    new(uint8(0x43)),
		MinimumAndMaximumChargingElectricPower: [2]*uint32{new(uint32(0)), nil},
		MinimumAndMaximumChargingCurrent:       [2]*uint16{nil, new(uint16(345))},
		RemainingCapacity2:                     new(uint16(10)),
		RemainingCapacity3:                     new(uint8(75)),
		ACChargeableElectricEnergy:             125, ACDischargeableElectricEnergy: 250,
		ACCumulativeChargingElectricEnergy: 500, ACCumulativeDischargingElectricEnergy: 750,
	}
	const want = `
# HELP echonetlite_storage_battery_ac_chargeable_electric_energy_joules AC chargeable electric energy. EPC=0xA4
# TYPE echonetlite_storage_battery_ac_chargeable_electric_energy_joules gauge
echonetlite_storage_battery_ac_chargeable_electric_energy_joules{eoj="027D01",host="192.0.2.10"} 450000
# HELP echonetlite_storage_battery_ac_charging_electric_energy_joules_total AC measured cumulative charging electric energy. EPC=0xA8
# TYPE echonetlite_storage_battery_ac_charging_electric_energy_joules_total counter
echonetlite_storage_battery_ac_charging_electric_energy_joules_total{eoj="027D01",host="192.0.2.10"} 1800000
# HELP echonetlite_storage_battery_ac_dischargeable_electric_energy_joules AC dischargeable electric energy. EPC=0xA5
# TYPE echonetlite_storage_battery_ac_dischargeable_electric_energy_joules gauge
echonetlite_storage_battery_ac_dischargeable_electric_energy_joules{eoj="027D01",host="192.0.2.10"} 900000
# HELP echonetlite_storage_battery_ac_discharging_electric_energy_joules_total AC measured cumulative discharging electric energy. EPC=0xA9
# TYPE echonetlite_storage_battery_ac_discharging_electric_energy_joules_total counter
echonetlite_storage_battery_ac_discharging_electric_energy_joules_total{eoj="027D01",host="192.0.2.10"} 2700000
# HELP echonetlite_storage_battery_actual_operation_mode Working operation status. values={"64":"Other","65":"Rapid charging","66":"Charging","67":"Discharging","68":"Standby","69":"Test","70":"Automatic","72":"Restart","73":"Effective capacity recalculation processing"} EPC=0xCF
# TYPE echonetlite_storage_battery_actual_operation_mode gauge
echonetlite_storage_battery_actual_operation_mode{eoj="027D01",host="192.0.2.10"} 67
# HELP echonetlite_storage_battery_minimum_and_maximum_charging_current_amperes Minimum/maximum charging current. EPC=0xCA
# TYPE echonetlite_storage_battery_minimum_and_maximum_charging_current_amperes gauge
echonetlite_storage_battery_minimum_and_maximum_charging_current_amperes{bound="max",eoj="027D01",host="192.0.2.10"} 34.5
# HELP echonetlite_storage_battery_minimum_and_maximum_charging_electric_power_watts Minimum/maximum charging electric power. EPC=0xC8
# TYPE echonetlite_storage_battery_minimum_and_maximum_charging_electric_power_watts gauge
echonetlite_storage_battery_minimum_and_maximum_charging_electric_power_watts{bound="min",eoj="027D01",host="192.0.2.10"} 0
# HELP echonetlite_storage_battery_remaining_capacity2_coulombs Remaining stored electricity 2. EPC=0xE3
# TYPE echonetlite_storage_battery_remaining_capacity2_coulombs gauge
echonetlite_storage_battery_remaining_capacity2_coulombs{eoj="027D01",host="192.0.2.10"} 3600
# HELP echonetlite_storage_battery_remaining_capacity3_ratio Remaining stored electricity 3. EPC=0xE4
# TYPE echonetlite_storage_battery_remaining_capacity3_ratio gauge
echonetlite_storage_battery_remaining_capacity3_ratio{eoj="027D01",host="192.0.2.10"} 0.75
# HELP echonetlite_storage_battery_fault_status Fault status. values={"65":"Fault occurred.","66":"No fault has occurred."} EPC=0x88
# TYPE echonetlite_storage_battery_fault_status gauge
echonetlite_storage_battery_fault_status{eoj="027D01",host="192.0.2.10"} 66
# HELP echonetlite_storage_battery_operation_status Operation status. values={"48":"ON","49":"OFF"} EPC=0x80
# TYPE echonetlite_storage_battery_operation_status gauge
echonetlite_storage_battery_operation_status{eoj="027D01",host="192.0.2.10"} 48
`
	assertPublishedMetrics(t, NewStorageBatteryCollector, echonetlite.ClassStorageBattery, reading, want)
}

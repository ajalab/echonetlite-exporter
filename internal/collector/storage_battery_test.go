package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestStorageBatteryCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.StorageBattery{
		ACChargeableElectricEnergy: 125, ACDischargeableElectricEnergy: 250,
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
`
	assertPublishedMetrics(t, NewStorageBatteryCollector, echonetlite.ClassStorageBattery, reading, want)
}

package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestDRElectricEnergyMeterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.DRElectricEnergyMeter{
		CumulativeAmountsOfElectricEnergyUnit: 0.1,
		ACInputCumulativeElectricEnergy:       new(uint32(10)), ACOutputCumulativeElectricEnergy: new(uint32(20)),
		IndependentOperationCumulativeElectricEnergy: new(uint32(1)),
		ACInstantaneousElectricPower:                 new(int32(-100)), IndependentOperationInstantaneousElectricPower: new(int32(55)),
	}
	const want = `
# HELP echonetlite_dr_electric_energy_meter_ac_electric_power_watts Measured instantaneous electric power (AC input/output).
# TYPE echonetlite_dr_electric_energy_meter_ac_electric_power_watts gauge
echonetlite_dr_electric_energy_meter_ac_electric_power_watts{eoj="028E01",host="192.0.2.10"} -100
# HELP echonetlite_dr_electric_energy_meter_ac_input_electric_energy_joules_total Measured cumulative amounts of electric energy (AC input).
# TYPE echonetlite_dr_electric_energy_meter_ac_input_electric_energy_joules_total counter
echonetlite_dr_electric_energy_meter_ac_input_electric_energy_joules_total{eoj="028E01",host="192.0.2.10"} 3600000
# HELP echonetlite_dr_electric_energy_meter_ac_output_electric_energy_joules_total Measured cumulative amounts of electric energy (AC output).
# TYPE echonetlite_dr_electric_energy_meter_ac_output_electric_energy_joules_total counter
echonetlite_dr_electric_energy_meter_ac_output_electric_energy_joules_total{eoj="028E01",host="192.0.2.10"} 7200000
# HELP echonetlite_dr_electric_energy_meter_independent_operation_electric_energy_joules_total Measured cumulative amounts of electric energy (output during a power outage).
# TYPE echonetlite_dr_electric_energy_meter_independent_operation_electric_energy_joules_total counter
echonetlite_dr_electric_energy_meter_independent_operation_electric_energy_joules_total{eoj="028E01",host="192.0.2.10"} 360000
# HELP echonetlite_dr_electric_energy_meter_independent_operation_electric_power_watts Measured instantaneous electric power (output during a power outage).
# TYPE echonetlite_dr_electric_energy_meter_independent_operation_electric_power_watts gauge
echonetlite_dr_electric_energy_meter_independent_operation_electric_power_watts{eoj="028E01",host="192.0.2.10"} 55
`
	assertPublishedMetrics(t, NewDRElectricEnergyMeterCollector, echonetlite.ClassDRElectricEnergyMeter, reading, want)
}

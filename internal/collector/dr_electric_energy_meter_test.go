package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestDRElectricEnergyMeterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.DRElectricEnergyMeter{
		SuperClass:                            echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		CumulativeAmountsOfElectricEnergyUnit: 0.1,
		ACInputCumulativeElectricEnergy:       new(uint32(10)), ACOutputCumulativeElectricEnergy: new(uint32(20)),
		IndependentOperationCumulativeElectricEnergy: new(uint32(1)),
		ACInstantaneousElectricPower:                 new(int32(-100)), IndependentOperationInstantaneousElectricPower: new(int32(55)),
	}
	const want = `
# HELP echonetlite_dr_electric_energy_meter_ac_electric_power_watts Measured instantaneous electric power (AC input/output). EPC=0xE9
# TYPE echonetlite_dr_electric_energy_meter_ac_electric_power_watts gauge
echonetlite_dr_electric_energy_meter_ac_electric_power_watts{eoj="028E01",host="192.0.2.10"} -100
# HELP echonetlite_dr_electric_energy_meter_ac_input_electric_energy_joules_total Measured cumulative amounts of electric energy (AC input). EPC=0xE0
# TYPE echonetlite_dr_electric_energy_meter_ac_input_electric_energy_joules_total counter
echonetlite_dr_electric_energy_meter_ac_input_electric_energy_joules_total{eoj="028E01",host="192.0.2.10"} 3600000
# HELP echonetlite_dr_electric_energy_meter_ac_output_electric_energy_joules_total Measured cumulative amounts of electric energy (AC output). EPC=0xE2
# TYPE echonetlite_dr_electric_energy_meter_ac_output_electric_energy_joules_total counter
echonetlite_dr_electric_energy_meter_ac_output_electric_energy_joules_total{eoj="028E01",host="192.0.2.10"} 7200000
# HELP echonetlite_dr_electric_energy_meter_independent_operation_electric_energy_joules_total Measured cumulative amounts of electric energy (output during a power outage). EPC=0xE4
# TYPE echonetlite_dr_electric_energy_meter_independent_operation_electric_energy_joules_total counter
echonetlite_dr_electric_energy_meter_independent_operation_electric_energy_joules_total{eoj="028E01",host="192.0.2.10"} 360000
# HELP echonetlite_dr_electric_energy_meter_independent_operation_electric_power_watts Measured instantaneous electric power (output during a power outage). EPC=0xEA
# TYPE echonetlite_dr_electric_energy_meter_independent_operation_electric_power_watts gauge
echonetlite_dr_electric_energy_meter_independent_operation_electric_power_watts{eoj="028E01",host="192.0.2.10"} 55
# HELP echonetlite_dr_electric_energy_meter_fault_status Fault status. values={"65":"Fault occurred.","66":"No fault has occurred."} EPC=0x88
# TYPE echonetlite_dr_electric_energy_meter_fault_status gauge
echonetlite_dr_electric_energy_meter_fault_status{eoj="028E01",host="192.0.2.10"} 66
# HELP echonetlite_dr_electric_energy_meter_operation_status Operation status. values={"48":"ON","49":"OFF"} EPC=0x80
# TYPE echonetlite_dr_electric_energy_meter_operation_status gauge
echonetlite_dr_electric_energy_meter_operation_status{eoj="028E01",host="192.0.2.10"} 48
`
	assertPublishedMetrics(t, NewDRElectricEnergyMeterCollector, echonetlite.ClassDRElectricEnergyMeter, reading, want)
}

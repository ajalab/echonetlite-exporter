package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestPVPowerGenerationCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.PVPowerGeneration{
		SuperClass:                           echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		InstantaneousElectricPowerGeneration: 450, CumulativeElectricEnergyOfGeneration: 240,
	}
	const want = `
# HELP echonetlite_pv_power_generation_electric_energy_of_generation_joules_total Measured cumulative amount of electric energy generated. EPC=0xE1
# TYPE echonetlite_pv_power_generation_electric_energy_of_generation_joules_total counter
echonetlite_pv_power_generation_electric_energy_of_generation_joules_total{eoj="027901",host="192.0.2.10"} 864000
# HELP echonetlite_pv_power_generation_electric_power_generation_watts Measured instantaneous amount of electricity generated. EPC=0xE0
# TYPE echonetlite_pv_power_generation_electric_power_generation_watts gauge
echonetlite_pv_power_generation_electric_power_generation_watts{eoj="027901",host="192.0.2.10"} 450
# HELP echonetlite_pv_power_generation_fault_status Fault status. EPC=0x88
# TYPE echonetlite_pv_power_generation_fault_status gauge
echonetlite_pv_power_generation_fault_status{eoj="027901",host="192.0.2.10"} 66
# HELP echonetlite_pv_power_generation_operation_status Operation status. EPC=0x80
# TYPE echonetlite_pv_power_generation_operation_status gauge
echonetlite_pv_power_generation_operation_status{eoj="027901",host="192.0.2.10"} 48
`
	assertPublishedMetrics(t, NewPVPowerGenerationCollector, echonetlite.ClassPVPowerGeneration, reading, want)
}

package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestMultipleInputPCSCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.MultipleInputPCS{
		SuperClass:                    echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		NormalDirectionElectricEnergy: 10, ReverseDirectionElectricEnergy: 20,
		InstantaneousElectricPower: -100,
	}
	const want = `
# HELP echonetlite_multiple_input_pcs_electric_power_watts Measured instantaneous electric power. EPC=0xE7
# TYPE echonetlite_multiple_input_pcs_electric_power_watts gauge
echonetlite_multiple_input_pcs_electric_power_watts{eoj="02A501",host="192.0.2.10"} -100
# HELP echonetlite_multiple_input_pcs_normal_direction_electric_energy_joules_total Measured cumulative amount of electric energy (normal direction). EPC=0xE0
# TYPE echonetlite_multiple_input_pcs_normal_direction_electric_energy_joules_total counter
echonetlite_multiple_input_pcs_normal_direction_electric_energy_joules_total{eoj="02A501",host="192.0.2.10"} 36000
# HELP echonetlite_multiple_input_pcs_reverse_direction_electric_energy_joules_total Measured cumulative amount of electric energy (reverse direction). EPC=0xE3
# TYPE echonetlite_multiple_input_pcs_reverse_direction_electric_energy_joules_total counter
echonetlite_multiple_input_pcs_reverse_direction_electric_energy_joules_total{eoj="02A501",host="192.0.2.10"} 72000
# HELP echonetlite_multiple_input_pcs_fault_status Fault status. values={"65":"Fault occurred.","66":"No fault has occurred."} EPC=0x88
# TYPE echonetlite_multiple_input_pcs_fault_status gauge
echonetlite_multiple_input_pcs_fault_status{eoj="02A501",host="192.0.2.10"} 66
# HELP echonetlite_multiple_input_pcs_operation_status Operation status. values={"48":"On","49":"Off"} EPC=0x80
# TYPE echonetlite_multiple_input_pcs_operation_status gauge
echonetlite_multiple_input_pcs_operation_status{eoj="02A501",host="192.0.2.10"} 48
`
	assertPublishedMetrics(t, NewMultipleInputPCSCollector, echonetlite.ClassMultipleInputPCS, reading, want)
}

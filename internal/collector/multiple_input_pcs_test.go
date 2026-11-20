package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestMultipleInputPCSCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.MultipleInputPCS{
		NormalDirectionElectricEnergy: 10, ReverseDirectionElectricEnergy: 20,
		InstantaneousElectricPower: -100,
	}
	const want = `
# HELP echonetlite_multiple_input_pcs_electric_power_watts Measured instantaneous electric power.
# TYPE echonetlite_multiple_input_pcs_electric_power_watts gauge
echonetlite_multiple_input_pcs_electric_power_watts{eoj="02A501",host="192.0.2.10"} -100
# HELP echonetlite_multiple_input_pcs_normal_direction_electric_energy_joules_total Measured cumulative amount of electric energy (normal direction).
# TYPE echonetlite_multiple_input_pcs_normal_direction_electric_energy_joules_total counter
echonetlite_multiple_input_pcs_normal_direction_electric_energy_joules_total{eoj="02A501",host="192.0.2.10"} 36000
# HELP echonetlite_multiple_input_pcs_reverse_direction_electric_energy_joules_total Measured cumulative amount of electric energy (reverse direction).
# TYPE echonetlite_multiple_input_pcs_reverse_direction_electric_energy_joules_total counter
echonetlite_multiple_input_pcs_reverse_direction_electric_energy_joules_total{eoj="02A501",host="192.0.2.10"} 72000
`
	assertPublishedMetrics(t, NewMultipleInputPCSCollector, echonetlite.ClassMultipleInputPCS, reading, want)
}

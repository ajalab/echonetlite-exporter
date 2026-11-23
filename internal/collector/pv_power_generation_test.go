package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestPVPowerGenerationCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.PVPowerGeneration{
		InstantaneousElectricPowerGeneration: 450, CumulativeElectricEnergyOfGeneration: 240,
	}
	const want = `
# HELP echonetlite_pv_power_generation_electric_energy_of_generation_joules_total Measured cumulative amount of electric energy generated. EPC=0xE1
# TYPE echonetlite_pv_power_generation_electric_energy_of_generation_joules_total counter
echonetlite_pv_power_generation_electric_energy_of_generation_joules_total{eoj="027901",host="192.0.2.10"} 864000
# HELP echonetlite_pv_power_generation_electric_power_generation_watts Measured instantaneous amount of electricity generated. EPC=0xE0
# TYPE echonetlite_pv_power_generation_electric_power_generation_watts gauge
echonetlite_pv_power_generation_electric_power_generation_watts{eoj="027901",host="192.0.2.10"} 450
`
	assertPublishedMetrics(t, NewPVPowerGenerationCollector, echonetlite.ClassPVPowerGeneration, reading, want)
}

package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestPowerDistributionBoardMeteringCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.PowerDistributionBoardMetering{
		UnitForCumulativeElectricEnergy: 0.1,
		InstantaneousElectricPowerListSimplex: echonetlite.InstantaneousElectricPowerListSimplex{
			StartChannel: 3, InstantaneousElectricPower: []*int32{new(int32(100)), nil, new(int32(-60))},
		},
		CumulativeElectricEnergyListSimplex: echonetlite.CumulativeElectricEnergyListSimplex{
			StartChannel: 3, ElectricEnergy: []*uint32{new(uint32(2))},
		},
	}
	const want = `
# HELP echonetlite_power_distribution_board_metering_electric_energy_simplex_joules_total Measured cumulative amount of electric power consumption list (simplex).
# TYPE echonetlite_power_distribution_board_metering_electric_energy_simplex_joules_total counter
echonetlite_power_distribution_board_metering_electric_energy_simplex_joules_total{channel="3",eoj="028701",host="192.0.2.10"} 720000
# HELP echonetlite_power_distribution_board_metering_electric_power_simplex_watts Measured instantaneous power consumption list (simplex).
# TYPE echonetlite_power_distribution_board_metering_electric_power_simplex_watts gauge
echonetlite_power_distribution_board_metering_electric_power_simplex_watts{channel="3",eoj="028701",host="192.0.2.10"} 100
echonetlite_power_distribution_board_metering_electric_power_simplex_watts{channel="5",eoj="028701",host="192.0.2.10"} -60
`
	assertPublishedMetrics(t, NewPowerDistributionBoardMeteringCollector, echonetlite.ClassPowerDistributionBoardMetering, reading, want)
}

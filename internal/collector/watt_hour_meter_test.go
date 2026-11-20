package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestWattHourMeterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.WattHourMeter{
		CumulativeElectricEnergy: 100, CumulativeAmountsOfElectricEnergyUnit: 0.01,
	}
	const want = `
# HELP echonetlite_watt_hour_meter_electric_energy_joules_total Cumulative amounts of electric energy measurement value.
# TYPE echonetlite_watt_hour_meter_electric_energy_joules_total counter
echonetlite_watt_hour_meter_electric_energy_joules_total{eoj="028001",host="192.0.2.10"} 3600000
`
	assertPublishedMetrics(t, NewWattHourMeterCollector, echonetlite.ClassWattHourMeter, reading, want)
}

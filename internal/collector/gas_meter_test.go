package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestGasMeterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.GasMeter{
		CumulativeGasConsumption: 1234,
	}
	const want = `
# HELP echonetlite_gas_meter_gas_consumption_cubic_meters_total Cumulative amount of gas consumption measurement value. EPC=0xE0
# TYPE echonetlite_gas_meter_gas_consumption_cubic_meters_total counter
echonetlite_gas_meter_gas_consumption_cubic_meters_total{eoj="028201",host="192.0.2.10"} 1.234
`
	assertPublishedMetrics(t, NewGasMeterCollector, echonetlite.ClassGasMeter, reading, want)
}

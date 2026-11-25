package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestGasMeterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.GasMeter{
		SuperClass:               echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		CumulativeGasConsumption: 1234,
	}
	const want = `
# HELP echonetlite_gas_meter_gas_consumption_cubic_meters_total Cumulative amount of gas consumption measurement value. EPC=0xE0
# TYPE echonetlite_gas_meter_gas_consumption_cubic_meters_total counter
echonetlite_gas_meter_gas_consumption_cubic_meters_total{eoj="028201",host="192.0.2.10"} 1.234
# HELP echonetlite_gas_meter_fault_status Fault status. EPC=0x88
# TYPE echonetlite_gas_meter_fault_status gauge
echonetlite_gas_meter_fault_status{eoj="028201",host="192.0.2.10"} 66
# HELP echonetlite_gas_meter_operation_status Operation status. EPC=0x80
# TYPE echonetlite_gas_meter_operation_status gauge
echonetlite_gas_meter_operation_status{eoj="028201",host="192.0.2.10"} 48
`
	assertPublishedMetrics(t, NewGasMeterCollector, echonetlite.ClassGasMeter, reading, want)
}

package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestWattHourMeterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.WattHourMeter{
		SuperClass:               echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		CumulativeElectricEnergy: 100, CumulativeAmountsOfElectricEnergyUnit: 0.01,
	}
	const want = `
# HELP echonetlite_watt_hour_meter_electric_energy_joules_total Cumulative amounts of electric energy measurement value. EPC=0xE0
# TYPE echonetlite_watt_hour_meter_electric_energy_joules_total counter
echonetlite_watt_hour_meter_electric_energy_joules_total{eoj="028001",host="192.0.2.10"} 3600000
# HELP echonetlite_watt_hour_meter_fault_status Fault status. values={"65":"Fault occurred.","66":"No fault has occurred."} EPC=0x88
# TYPE echonetlite_watt_hour_meter_fault_status gauge
echonetlite_watt_hour_meter_fault_status{eoj="028001",host="192.0.2.10"} 66
# HELP echonetlite_watt_hour_meter_operation_status Operation status. values={"48":"ON","49":"OFF"} EPC=0x80
# TYPE echonetlite_watt_hour_meter_operation_status gauge
echonetlite_watt_hour_meter_operation_status{eoj="028001",host="192.0.2.10"} 48
`
	assertPublishedMetrics(t, NewWattHourMeterCollector, echonetlite.ClassWattHourMeter, reading, want)
}

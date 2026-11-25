package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestWaterFlowMeterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.WaterFlowMeter{

		SuperClass:              echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		AbnormalDetection:       new(uint8(0x41)),
		SecurityDataInformation: new(uint32(0xFFFFFFFF)),
		CumulativeFlowingWater:  100, UnitForCumulativeFlowingWater: 0.001,
	}
	const want = `
# HELP echonetlite_water_flow_meter_flowing_water_cubic_meters_total Measured cumulative amount of flowing water. EPC=0xE0
# TYPE echonetlite_water_flow_meter_flowing_water_cubic_meters_total counter
echonetlite_water_flow_meter_flowing_water_cubic_meters_total{eoj="028101",host="192.0.2.10"} 0.1
# HELP echonetlite_water_flow_meter_abnormal_detection Detection of abnormal value in metering data. EPC=0xE3
# TYPE echonetlite_water_flow_meter_abnormal_detection gauge
echonetlite_water_flow_meter_abnormal_detection{eoj="028101",host="192.0.2.10"} 65
# HELP echonetlite_water_flow_meter_fault_status Fault status. EPC=0x88
# TYPE echonetlite_water_flow_meter_fault_status gauge
echonetlite_water_flow_meter_fault_status{eoj="028101",host="192.0.2.10"} 66
# HELP echonetlite_water_flow_meter_operation_status Operation status. EPC=0x80
# TYPE echonetlite_water_flow_meter_operation_status gauge
echonetlite_water_flow_meter_operation_status{eoj="028101",host="192.0.2.10"} 48
# HELP echonetlite_water_flow_meter_security_data_information Security data information. EPC=0xE4
# TYPE echonetlite_water_flow_meter_security_data_information gauge
echonetlite_water_flow_meter_security_data_information{eoj="028101",host="192.0.2.10"} 4294967295
`
	assertPublishedMetrics(t, NewWaterFlowMeterCollector, echonetlite.ClassWaterFlowMeter, reading, want)
}

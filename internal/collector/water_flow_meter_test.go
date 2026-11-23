package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestWaterFlowMeterCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.WaterFlowMeter{
		CumulativeFlowingWater: 100, UnitForCumulativeFlowingWater: 0.001,
	}
	const want = `
# HELP echonetlite_water_flow_meter_flowing_water_cubic_meters_total Measured cumulative amount of flowing water. EPC=0xE0
# TYPE echonetlite_water_flow_meter_flowing_water_cubic_meters_total counter
echonetlite_water_flow_meter_flowing_water_cubic_meters_total{eoj="028101",host="192.0.2.10"} 0.1
`
	assertPublishedMetrics(t, NewWaterFlowMeterCollector, echonetlite.ClassWaterFlowMeter, reading, want)
}

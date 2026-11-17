package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeWaterFlowMeterClient func(context.Context, string, echonetlite.EOJ) (*echonetlite.WaterFlowMeter, error)

func (f fakeWaterFlowMeterClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.WaterFlowMeter, error) {
	return f(ctx, host, eoj)
}

func TestWaterFlowMeterCollectorPublishesMetrics(t *testing.T) {
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x81, 0x01})
	targets := []echonetlite.Device{device}
	success := NewCollectMetrics()
	c := NewWaterFlowMeterCollector(
		fakeWaterFlowMeterClient(func(context.Context, string, echonetlite.EOJ) (*echonetlite.WaterFlowMeter, error) {
			return &echonetlite.WaterFlowMeter{
				CumulativeFlowingWater: 100, UnitForCumulativeFlowingWater: 0.001,
			}, nil
		}),
		time.Second, time.Second, success, targets,
	)
	c.collect(t.Context(), targets)
	assertCollectSuccess(t, success, device, 1)

	const want = `
# HELP echonetlite_water_flow_meter_flowing_water_cubic_meters_total Measured cumulative amount of flowing water.
# TYPE echonetlite_water_flow_meter_flowing_water_cubic_meters_total counter
echonetlite_water_flow_meter_flowing_water_cubic_meters_total{eoj="028101",host="192.0.2.10"} 0.1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	}
}

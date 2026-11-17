package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeGasMeterClient func(context.Context, string, echonetlite.EOJ) (*echonetlite.GasMeter, error)

func (f fakeGasMeterClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.GasMeter, error) {
	return f(ctx, host, eoj)
}

func TestGasMeterCollectorPublishesMetrics(t *testing.T) {
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x82, 0x01})
	targets := []echonetlite.Device{device}
	success := NewCollectMetrics()
	c := NewGasMeterCollector(
		fakeGasMeterClient(func(context.Context, string, echonetlite.EOJ) (*echonetlite.GasMeter, error) {
			return &echonetlite.GasMeter{
				CumulativeGasConsumption: 100,
			}, nil
		}),
		time.Second, time.Second, success, targets,
	)
	c.collect(t.Context(), targets)
	assertCollectSuccess(t, success, device, 1)

	const want = `
# HELP echonetlite_gas_meter_gas_consumption_cubic_meters_total Cumulative amount of gas consumption measurement value.
# TYPE echonetlite_gas_meter_gas_consumption_cubic_meters_total counter
echonetlite_gas_meter_gas_consumption_cubic_meters_total{eoj="028201",host="192.0.2.10"} 0.1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	}
}

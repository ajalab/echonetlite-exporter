package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeWattHourMeterClient func(context.Context, string, echonetlite.EOJ) (*echonetlite.WattHourMeter, error)

func (f fakeWattHourMeterClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.WattHourMeter, error) {
	return f(ctx, host, eoj)
}

func TestWattHourMeterCollectorPublishesMetrics(t *testing.T) {
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x80, 0x01})
	targets := []echonetlite.Device{device}
	success := NewCollectMetrics()
	c := NewWattHourMeterCollector(
		fakeWattHourMeterClient(func(context.Context, string, echonetlite.EOJ) (*echonetlite.WattHourMeter, error) {
			return &echonetlite.WattHourMeter{
				CumulativeElectricEnergy: 100, CumulativeAmountsOfElectricEnergyUnit: 0.01,
			}, nil
		}),
		time.Second, time.Second, success, targets,
	)
	c.collect(t.Context(), targets)
	assertCollectSuccess(t, success, device, 1)

	const want = `
# HELP echonetlite_watt_hour_meter_electric_energy_joules_total Cumulative amounts of electric energy measurement value.
# TYPE echonetlite_watt_hour_meter_electric_energy_joules_total counter
echonetlite_watt_hour_meter_electric_energy_joules_total{eoj="028001",host="192.0.2.10"} 3600000
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	}
}

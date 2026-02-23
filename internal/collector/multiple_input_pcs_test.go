package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMultipleInputPCSCollectSuccessUpdatesMetricsAndSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.20", echonetlite.EOJ{0x02, 0xA5, 0x01})
	c := NewMultipleInputPCSCollector(nil, time.Second, time.Second, collectMetrics, []echonetlite.Device{device})

	c.get = func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.MultipleInputPCS, error) {
		return &echonetlite.MultipleInputPCS{
			NormalDirectionElectricEnergy:  2,
			ReverseDirectionElectricEnergy: 3,
			InstantaneousElectricPower:     -120,
		}, nil
	}
	c.collect(context.Background(), []echonetlite.Device{device})
	assertCollectSuccess(t, collectMetrics, device, 1)

	gotPower := testutil.ToFloat64(
		c.instantaneousElectricPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	assertFloatEqual(t, gotPower, -120)

	key := cumulativeMultipleInputPCSEnergyKey{host: device.Host(), eoj: device.EOJ().String()}
	gotNormal, ok := c.normalDirectionEnergy[key]
	if !ok {
		t.Fatalf("expected normal direction energy for key %+v", key)
	}
	assertFloatEqual(t, gotNormal, 7200)
	gotReverse, ok := c.reverseDirectionEnergy[key]
	if !ok {
		t.Fatalf("expected reverse direction energy for key %+v", key)
	}
	assertFloatEqual(t, gotReverse, 10800)
}

func TestMultipleInputPCSCollectFailureSetsSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.20", echonetlite.EOJ{0x02, 0xA5, 0x01})
	c := NewMultipleInputPCSCollector(nil, time.Second, time.Second, collectMetrics, []echonetlite.Device{device})

	c.get = func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.MultipleInputPCS, error) {
		return nil, errors.New("test error")
	}
	c.collect(context.Background(), []echonetlite.Device{device})
	assertCollectSuccess(t, collectMetrics, device, 0)
}

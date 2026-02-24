package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakePVPowerGenerationClient struct {
	getFn func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.PVPowerGeneration, error)
}

func (f fakePVPowerGenerationClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.PVPowerGeneration, error) {
	return f.getFn(ctx, host, eoj)
}

func TestPVPowerGenerationCollectSuccessUpdatesMetricsAndSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x79, 0x01})
	c := NewPVPowerGenerationCollector(
		fakePVPowerGenerationClient{
			getFn: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.PVPowerGeneration, error) {
				return &echonetlite.PVPowerGeneration{
					InstantaneousElectricPowerGeneration: 321,
					CumulativeElectricEnergyOfGeneration: 2000, // 2kWh
				}, nil
			},
		},
		time.Second,
		time.Second,
		collectMetrics,
		[]echonetlite.Device{device},
	)
	c.collect(context.Background(), []echonetlite.Device{device})
	assertCollectSuccess(t, collectMetrics, device, 1)

	gotGauge := testutil.ToFloat64(
		c.instantaneousPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	assertFloatEqual(t, gotGauge, 321)

	key := cumulativeEnergyKey{host: device.Host(), eoj: device.EOJ().String()}
	gotCounter, ok := c.cumulativeEnergy[key]
	if !ok {
		t.Fatalf("expected cumulative energy for key %+v", key)
	}
	assertFloatEqual(t, gotCounter, 7200000)
}

func TestPVPowerGenerationCollectFailureSetsSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x79, 0x01})
	c := NewPVPowerGenerationCollector(
		fakePVPowerGenerationClient{
			getFn: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.PVPowerGeneration, error) {
				return nil, errors.New("test error")
			},
		},
		time.Second,
		time.Second,
		collectMetrics,
		[]echonetlite.Device{device},
	)
	c.collect(context.Background(), []echonetlite.Device{device})
	assertCollectSuccess(t, collectMetrics, device, 0)
}

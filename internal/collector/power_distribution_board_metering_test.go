package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakePowerDistributionBoardMeteringClient struct {
	getFn func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.PowerDistributionBoardMetering, error)
}

func (f fakePowerDistributionBoardMeteringClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.PowerDistributionBoardMetering, error) {
	return f.getFn(ctx, host, eoj)
}

func TestPowerDistributionBoardMeteringCollectSuccessUpdatesMetricsAndSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.11", echonetlite.EOJ{0x02, 0x87, 0x01})
	c := NewPowerDistributionBoardMeteringCollector(
		fakePowerDistributionBoardMeteringClient{
			getFn: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.PowerDistributionBoardMetering, error) {
				return &echonetlite.PowerDistributionBoardMetering{
					InstantaneousElectricPowerListSimplex: echonetlite.InstantaneousElectricPowerListSimplex{
						StartChannel:               3,
						InstantaneousElectricPower: []int32{100},
					},
					CumulativeElectricEnergyListSimplex: echonetlite.CumulativeElectricEnergyListSimplex{
						StartChannel:   3,
						ElectricEnergy: []int32{2},
					},
					UnitForCumulativeEnergy: 0.1, // 0.2kWh
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
		c.instantaneousElectricPowerSimplexGauge.WithLabelValues(device.Host(), device.EOJ().String(), "3"),
	)
	assertFloatEqual(t, gotGauge, 100)

	key := cumulativeElectricEnergyKey{
		host:    device.Host(),
		eoj:     device.EOJ().String(),
		channel: "3",
	}
	gotCounter, ok := c.cumulativeElectricEnergySimplex[key]
	if !ok {
		t.Fatalf("expected cumulative energy for key %+v", key)
	}
	assertFloatApprox(t, gotCounter, 720000, 0.1)
}

func TestPowerDistributionBoardMeteringCollectFailureSetsSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.11", echonetlite.EOJ{0x02, 0x87, 0x01})
	c := NewPowerDistributionBoardMeteringCollector(
		fakePowerDistributionBoardMeteringClient{
			getFn: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.PowerDistributionBoardMetering, error) {
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

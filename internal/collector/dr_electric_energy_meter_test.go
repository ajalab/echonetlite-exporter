package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeDRElectricEnergyMeterClient struct {
	getFn func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.DRElectricEnergyMeter, error)
}

func (f fakeDRElectricEnergyMeterClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.DRElectricEnergyMeter, error) {
	return f.getFn(ctx, host, eoj)
}

func TestDRElectricEnergyMeterCollectSuccessUpdatesMetricsAndSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.20", echonetlite.EOJ{0x02, 0x8E, 0x01})
	c := NewDRElectricEnergyMeterCollector(
		fakeDRElectricEnergyMeterClient{
			getFn: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.DRElectricEnergyMeter, error) {
				return &echonetlite.DRElectricEnergyMeter{
					CumulativeAmountsOfElectricEnergyUnit:          0.01,
					ACInputCumulativeElectricEnergy:                10,
					ACOutputCumulativeElectricEnergy:               20,
					IndependentOperationCumulativeElectricEnergy:   30,
					ACInstantaneousElectricPower:                   -120,
					IndependentOperationInstantaneousElectricPower: 80,
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

	gotACInOut := testutil.ToFloat64(c.acInOutInstantaneousPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()))
	assertFloatEqual(t, gotACInOut, -120)
	gotIndepPower := testutil.ToFloat64(c.independentOperationInstantaneousPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()))
	assertFloatEqual(t, gotIndepPower, 80)

	key := cumulativeDRElectricEnergyMeterKey{host: device.Host(), eoj: device.EOJ().String()}
	gotACInput, ok := c.acInputCumulativeEnergy[key]
	if !ok {
		t.Fatalf("expected AC input cumulative energy for key %+v", key)
	}
	assertFloatApprox(t, gotACInput, 360000, 0.1)
	gotACOutput, ok := c.acOutputCumulativeEnergy[key]
	if !ok {
		t.Fatalf("expected AC output cumulative energy for key %+v", key)
	}
	assertFloatApprox(t, gotACOutput, 720000, 0.1)
	gotIndepEnergy, ok := c.independentOperationCumulativeEnergy[key]
	if !ok {
		t.Fatalf("expected independent operation cumulative energy for key %+v", key)
	}
	assertFloatApprox(t, gotIndepEnergy, 1080000, 0.1)
}

func TestDRElectricEnergyMeterCollectFailureSetsSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.20", echonetlite.EOJ{0x02, 0x8E, 0x01})
	c := NewDRElectricEnergyMeterCollector(
		fakeDRElectricEnergyMeterClient{
			getFn: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.DRElectricEnergyMeter, error) {
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

package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type stubMultipleInputPCSClient struct {
	getFunc func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.MultipleInputPCS, error)
}

func (s *stubMultipleInputPCSClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.MultipleInputPCS, error) {
	return s.getFunc(ctx, host, eoj)
}

func TestMultipleInputPCSUpdateMetricsUsesDeviceLabels(t *testing.T) {
	c := NewMultipleInputPCSCollector(nil, 0, 0, NewCollectMetrics(), nil)
	device := echonetlite.NewDevice("192.0.2.20", echonetlite.EOJ{0x02, 0xA5, 0x01})
	mipcs := &echonetlite.MultipleInputPCS{
		NormalDirectionElectricEnergy:  2,
		ReverseDirectionElectricEnergy: 3,
		InstantaneousElectricPower:     -120,
	}

	c.updateMetrics(device, mipcs)

	gotPower := testutil.ToFloat64(
		c.instantaneousElectricPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	if gotPower != -120 {
		t.Fatalf("expected instantaneous power -120, got %v", gotPower)
	}

	key := cumulativeMultipleInputPCSEnergyKey{host: device.Host(), eoj: device.EOJ().String()}
	gotNormal, ok := c.normalDirectionEnergy[key]
	if !ok {
		t.Fatalf("expected normal direction energy for key %+v", key)
	}
	if gotNormal != 7200 {
		t.Fatalf("expected normal direction energy 7200, got %v", gotNormal)
	}
	gotReverse, ok := c.reverseDirectionEnergy[key]
	if !ok {
		t.Fatalf("expected reverse direction energy for key %+v", key)
	}
	if gotReverse != 10800 {
		t.Fatalf("expected reverse direction energy 10800, got %v", gotReverse)
	}
}

func TestMultipleInputPCSCollectSetsSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.20", echonetlite.EOJ{0x02, 0xA5, 0x01})
	c := NewMultipleInputPCSCollector(nil, time.Second, time.Second, collectMetrics, []echonetlite.Device{device})

	c.client = &stubMultipleInputPCSClient{
		getFunc: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.MultipleInputPCS, error) {
			return nil, errors.New("test error")
		},
	}
	c.collect(context.Background(), []echonetlite.Device{device})
	gotFail := testutil.ToFloat64(
		collectMetrics.successGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	if gotFail != 0 {
		t.Fatalf("expected success gauge 0 after failure, got %v", gotFail)
	}

	c.client = &stubMultipleInputPCSClient{
		getFunc: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.MultipleInputPCS, error) {
			return &echonetlite.MultipleInputPCS{
				NormalDirectionElectricEnergy:  1,
				ReverseDirectionElectricEnergy: 2,
				InstantaneousElectricPower:     -3,
			}, nil
		},
	}
	c.collect(context.Background(), []echonetlite.Device{device})
	gotSuccess := testutil.ToFloat64(
		collectMetrics.successGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	if gotSuccess != 1 {
		t.Fatalf("expected success gauge 1 after success, got %v", gotSuccess)
	}
}

package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type stubStorageBatteryClient struct {
	getFunc func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.StorageBattery, error)
}

func (s *stubStorageBatteryClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.StorageBattery, error) {
	return s.getFunc(ctx, host, eoj)
}

func TestStorageBatteryUpdateMetricsUsesDeviceLabels(t *testing.T) {
	c := NewStorageBatteryCollector(nil, 0, 0, NewCollectMetrics(), nil)
	device := echonetlite.NewDevice("192.0.2.12", echonetlite.EOJ{0x02, 0x7D, 0x01})
	sb := &echonetlite.StorageBattery{
		ACChargeableElectricEnergy:            2,
		ACDischargeableElectricEnergy:         3,
		ACCumulativeChargingElectricEnergy:    4,
		ACCumulativeDischargingElectricEnergy: 5,
	}

	c.updateMetrics(device, sb)

	gotChargeable := testutil.ToFloat64(
		c.acChargeableEnergyGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	if gotChargeable != 7200 {
		t.Fatalf("expected chargeable energy 7200, got %v", gotChargeable)
	}
	gotDischargeable := testutil.ToFloat64(
		c.acDischargeableEnergyGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	if gotDischargeable != 10800 {
		t.Fatalf("expected dischargeable energy 10800, got %v", gotDischargeable)
	}

	key := cumulativeStorageBatteryEnergyKey{host: device.Host(), eoj: device.EOJ().String()}
	gotCharging, ok := c.acChargingEnergy[key]
	if !ok {
		t.Fatalf("expected cumulative charging energy for key %+v", key)
	}
	if gotCharging != 14400 {
		t.Fatalf("expected cumulative charging energy 14400, got %v", gotCharging)
	}
	gotDischarging, ok := c.acDischargingEnergy[key]
	if !ok {
		t.Fatalf("expected cumulative discharging energy for key %+v", key)
	}
	if gotDischarging != 18000 {
		t.Fatalf("expected cumulative discharging energy 18000, got %v", gotDischarging)
	}
}

func TestStorageBatteryCollectSetsSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.12", echonetlite.EOJ{0x02, 0x7D, 0x01})
	c := NewStorageBatteryCollector(nil, time.Second, time.Second, collectMetrics, []echonetlite.Device{device})

	c.client = &stubStorageBatteryClient{
		getFunc: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.StorageBattery, error) {
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

	c.client = &stubStorageBatteryClient{
		getFunc: func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.StorageBattery, error) {
			return &echonetlite.StorageBattery{
				ACChargeableElectricEnergy:            1,
				ACDischargeableElectricEnergy:         2,
				ACCumulativeChargingElectricEnergy:    3,
				ACCumulativeDischargingElectricEnergy: 4,
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

package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestStorageBatteryCollectSuccessUpdatesMetricsAndSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.12", echonetlite.EOJ{0x02, 0x7D, 0x01})
	c := NewStorageBatteryCollector(nil, time.Second, time.Second, collectMetrics, []echonetlite.Device{device})

	c.get = func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.StorageBattery, error) {
		return &echonetlite.StorageBattery{
			ACChargeableElectricEnergy:            2,
			ACDischargeableElectricEnergy:         3,
			ACCumulativeChargingElectricEnergy:    4,
			ACCumulativeDischargingElectricEnergy: 5,
		}, nil
	}
	c.collect(context.Background(), []echonetlite.Device{device})
	assertCollectSuccess(t, collectMetrics, device, 1)

	gotChargeable := testutil.ToFloat64(
		c.acChargeableEnergyGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	assertFloatEqual(t, gotChargeable, 7200)
	gotDischargeable := testutil.ToFloat64(
		c.acDischargeableEnergyGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	assertFloatEqual(t, gotDischargeable, 10800)

	key := cumulativeStorageBatteryEnergyKey{host: device.Host(), eoj: device.EOJ().String()}
	gotCharging, ok := c.acChargingEnergy[key]
	if !ok {
		t.Fatalf("expected cumulative charging energy for key %+v", key)
	}
	assertFloatEqual(t, gotCharging, 14400)
	gotDischarging, ok := c.acDischargingEnergy[key]
	if !ok {
		t.Fatalf("expected cumulative discharging energy for key %+v", key)
	}
	assertFloatEqual(t, gotDischarging, 18000)
}

func TestStorageBatteryCollectFailureSetsSuccessState(t *testing.T) {
	collectMetrics := NewCollectMetrics()
	device := echonetlite.NewDevice("192.0.2.12", echonetlite.EOJ{0x02, 0x7D, 0x01})
	c := NewStorageBatteryCollector(nil, time.Second, time.Second, collectMetrics, []echonetlite.Device{device})

	c.get = func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.StorageBattery, error) {
		return nil, errors.New("test error")
	}
	c.collect(context.Background(), []echonetlite.Device{device})
	assertCollectSuccess(t, collectMetrics, device, 0)
}

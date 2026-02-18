package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPVUpdateMetricsUsesDeviceLabels(t *testing.T) {
	c := NewPVPowerGenerationCollector(nil, 0, 0, NewCollectMetrics(), nil)
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x79, 0x01})
	pvpg := &echonetlite.PVPowerGeneration{
		InstantaneousElectricPowerGeneration: 321,
		CumulativeElectricEnergyOfGeneration: 2000, // 2kWh
	}

	c.updateMetrics(device, pvpg)

	gotGauge := testutil.ToFloat64(
		c.instantaneousPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	if gotGauge != 321 {
		t.Fatalf("expected instantaneous gauge 321, got %v", gotGauge)
	}

	key := cumulativeEnergyKey{host: device.Host(), eoj: device.EOJ().String()}
	gotCounter, ok := c.cumulativeEnergy[key]
	if !ok {
		t.Fatalf("expected cumulative energy for key %+v", key)
	}
	if gotCounter != 7200000 {
		t.Fatalf("expected cumulative energy 7200000, got %v", gotCounter)
	}
}

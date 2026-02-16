package collector

import (
	"math"
	"testing"

	"echonetlite-exporter/echonetlite"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPDBMUpdateMetricsUsesDeviceLabels(t *testing.T) {
	c := NewPowerDistributionBoardMeteringCollector(nil, 0, 0, NewCollectMetrics(), nil)
	device := echonetlite.NewDevice("192.0.2.11", echonetlite.EOJ{0x02, 0x87, 0x01})
	pdbm := &echonetlite.PowerDistributionBoardMetering{
		InstantaneousElectricPowerListSimplex: echonetlite.InstantaneousElectricPowerListSimplex{
			StartChannel:               3,
			InstantaneousElectricPower: []int32{100},
		},
		CumulativeElectricEnergyListSimplex: echonetlite.CumulativeElectricEnergyListSimplex{
			StartChannel:   3,
			ElectricEnergy: []int32{2},
		},
		UnitForCumulativeEnergy: 0.1, // 0.2kWh
	}

	c.updateMetrics(device, pdbm)

	gotGauge := testutil.ToFloat64(
		c.instantaneousElectricPowerSimplexGauge.WithLabelValues(device.Host(), device.EOJ().String(), "3"),
	)
	if gotGauge != 100 {
		t.Fatalf("expected instantaneous gauge 100, got %v", gotGauge)
	}

	key := cumulativeElectricEnergyKey{
		host:    device.Host(),
		eoj:     device.EOJ().String(),
		channel: "3",
	}
	gotCounter, ok := c.cumulativeElectricEnergySimplex[key]
	if !ok {
		t.Fatalf("expected cumulative energy for key %+v", key)
	}
	if math.Abs(gotCounter-720000) > 0.1 {
		t.Fatalf("expected cumulative energy about 720000, got %v", gotCounter)
	}
}

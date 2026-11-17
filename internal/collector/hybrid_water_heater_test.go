package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeHybridWaterHeaterClient func(context.Context, string, echonetlite.EOJ) (*echonetlite.HybridWaterHeater, error)

func (f fakeHybridWaterHeaterClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.HybridWaterHeater, error) {
	return f(ctx, host, eoj)
}

func TestHybridWaterHeaterCollectorPublishesMetrics(t *testing.T) {
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0xA6, 0x01})
	targets := []echonetlite.Device{device}
	enabled, disabled := true, false
	remainingWater, tankCapacity := uint16(150), uint16(200)
	success := NewCollectMetrics()
	c := NewHybridWaterHeaterCollector(
		fakeHybridWaterHeaterClient(func(context.Context, string, echonetlite.EOJ) (*echonetlite.HybridWaterHeater, error) {
			return &echonetlite.HybridWaterHeater{
				LinkageModeForSolarPowerGeneration: 0x42,
				WaterHeatingStatus:                 &enabled,
				HeaterStatus:                       &disabled,
				HotWaterSupplyStatus:               &enabled,
				RemainingWater:                     &remainingWater,
				TankCapacity:                       &tankCapacity,
			}, nil
		}),
		time.Second, time.Second, success, targets,
	)
	c.collect(t.Context(), targets)
	assertCollectSuccess(t, success, device, 1)

	const want = `
# HELP echonetlite_hybrid_water_heater_heater_status Heater status.
# TYPE echonetlite_hybrid_water_heater_heater_status gauge
echonetlite_hybrid_water_heater_heater_status{eoj="02A601",host="192.0.2.10"} 0
# HELP echonetlite_hybrid_water_heater_hot_water_supply_status Hot water supply status.
# TYPE echonetlite_hybrid_water_heater_hot_water_supply_status gauge
echonetlite_hybrid_water_heater_hot_water_supply_status{eoj="02A601",host="192.0.2.10"} 1
# HELP echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation Linkage mode setting for solar power generation.
# TYPE echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation gauge
echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation{eoj="02A601",host="192.0.2.10"} 66
# HELP echonetlite_hybrid_water_heater_remaining_water_cubic_meters Measured amount of hot water remaining in tank.
# TYPE echonetlite_hybrid_water_heater_remaining_water_cubic_meters gauge
echonetlite_hybrid_water_heater_remaining_water_cubic_meters{eoj="02A601",host="192.0.2.10"} 0.15
# HELP echonetlite_hybrid_water_heater_tank_capacity_cubic_meters Tank capacity.
# TYPE echonetlite_hybrid_water_heater_tank_capacity_cubic_meters gauge
echonetlite_hybrid_water_heater_tank_capacity_cubic_meters{eoj="02A601",host="192.0.2.10"} 0.2
# HELP echonetlite_hybrid_water_heater_water_heating_status Water heating status.
# TYPE echonetlite_hybrid_water_heater_water_heating_status gauge
echonetlite_hybrid_water_heater_water_heating_status{eoj="02A601",host="192.0.2.10"} 1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	}
}

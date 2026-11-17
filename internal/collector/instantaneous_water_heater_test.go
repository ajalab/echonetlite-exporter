package collector

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeInstantaneousWaterHeaterClient func(context.Context, string, echonetlite.EOJ) (*echonetlite.InstantaneousWaterHeater, error)

func (f fakeInstantaneousWaterHeaterClient) Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.InstantaneousWaterHeater, error) {
	return f(ctx, host, eoj)
}

func TestInstantaneousWaterHeaterCollectorPublishesMetrics(t *testing.T) {
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x72, 0x01})
	targets := []echonetlite.Device{device}
	enabled, disabled := true, false
	supplyTemperature, bathTemperature, bathState := uint8(42), uint8(40), uint8(0x43)
	success := NewCollectMetrics()
	c := NewInstantaneousWaterHeaterCollector(
		fakeInstantaneousWaterHeaterClient(func(context.Context, string, echonetlite.EOJ) (*echonetlite.InstantaneousWaterHeater, error) {
			return &echonetlite.InstantaneousWaterHeater{
				HotWaterHeatingStatus:          true,
				AutomaticBathOperation:         &disabled,
				TargetSuppliedWaterTemperature: &supplyTemperature,
				TargetBathWaterTemperature:     &bathTemperature,
				ShowerHotWaterSupplyStatus:     &enabled,
				KitchenHotWaterSupplyStatus:    &disabled,
				BathOperationStatus:            &bathState,
			}, nil
		}),
		time.Second, time.Second, success, targets,
	)
	c.collect(t.Context(), targets)
	assertCollectSuccess(t, success, device, 1)

	const want = `
# HELP echonetlite_instantaneous_water_heater_automatic_bath_operation Bath Auto mode setting.
# TYPE echonetlite_instantaneous_water_heater_automatic_bath_operation gauge
echonetlite_instantaneous_water_heater_automatic_bath_operation{eoj="027201",host="192.0.2.10"} 0
# HELP echonetlite_instantaneous_water_heater_bath_operation_status Bath operation status monitor.
# TYPE echonetlite_instantaneous_water_heater_bath_operation_status gauge
echonetlite_instantaneous_water_heater_bath_operation_status{eoj="027201",host="192.0.2.10"} 67
# HELP echonetlite_instantaneous_water_heater_bath_water_heating_status Bath water heater status.
# TYPE echonetlite_instantaneous_water_heater_bath_water_heating_status gauge
echonetlite_instantaneous_water_heater_bath_water_heating_status{eoj="027201",host="192.0.2.10"} 0
# HELP echonetlite_instantaneous_water_heater_hot_water_heating_status Hot water heating status.
# TYPE echonetlite_instantaneous_water_heater_hot_water_heating_status gauge
echonetlite_instantaneous_water_heater_hot_water_heating_status{eoj="027201",host="192.0.2.10"} 1
# HELP echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status Kitchen hot water supply status.
# TYPE echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status gauge
echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status{eoj="027201",host="192.0.2.10"} 0
# HELP echonetlite_instantaneous_water_heater_shower_hot_water_supply_status Shower hot water supply status.
# TYPE echonetlite_instantaneous_water_heater_shower_hot_water_supply_status gauge
echonetlite_instantaneous_water_heater_shower_hot_water_supply_status{eoj="027201",host="192.0.2.10"} 1
# HELP echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius Set value of bath temperature.
# TYPE echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius gauge
echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius{eoj="027201",host="192.0.2.10"} 40
# HELP echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius Set value of hot water temperature.
# TYPE echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius gauge
echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius{eoj="027201",host="192.0.2.10"} 42
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	}
}

func TestInstantaneousWaterHeaterCollectorReplacesOptionalSeriesAfterSuccessfulRead(t *testing.T) {
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x72, 0x01})
	success := NewCollectMetrics()
	const temperature = "echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius"
	zero := uint8(0)
	reading := &echonetlite.InstantaneousWaterHeater{TargetSuppliedWaterTemperature: &zero}
	var readErr error
	collector := NewInstantaneousWaterHeaterCollector(
		fakeInstantaneousWaterHeaterClient(func(context.Context, string, echonetlite.EOJ) (*echonetlite.InstantaneousWaterHeater, error) {
			return reading, readErr
		}), time.Second, time.Second, success, []echonetlite.Device{device},
	)

	collector.collect(t.Context(), []echonetlite.Device{device})
	assertFloatEqual(t, float64(testutil.CollectAndCount(collector, temperature)), 1)

	reading, readErr = nil, errors.New("required property unavailable")
	collector.collect(t.Context(), []echonetlite.Device{device})
	assertCollectSuccess(t, success, device, 0)
	assertFloatEqual(t, float64(testutil.CollectAndCount(collector, temperature)), 1)

	reading, readErr = &echonetlite.InstantaneousWaterHeater{}, nil
	collector.collect(t.Context(), []echonetlite.Device{device})
	assertCollectSuccess(t, success, device, 1)
	assertFloatEqual(t, float64(testutil.CollectAndCount(collector, temperature)), 0)
}

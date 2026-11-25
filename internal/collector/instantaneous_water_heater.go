package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// InstantaneousWaterHeaterCollector collects and publishes metrics for its ECHONET Lite device class.
type InstantaneousWaterHeaterCollector struct {
	*collector
}

// NewInstantaneousWaterHeaterCollector creates a InstantaneousWaterHeaterCollector.
func NewInstantaneousWaterHeaterCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.InstantaneousWaterHeater, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *InstantaneousWaterHeaterCollector {
	hotWaterHeating := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCHotWaterHeatingStatus,
		"hotWaterHeatingStatus",
		"echonetlite_instantaneous_water_heater_hot_water_heating_status",
		"Hot water heating status",
		`{"65":"Hot water heating status found","66":"Hot water heating status not found"}`,
	)
	bathHeating := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCBathWaterHeatingStatus,
		"bathWaterHeatingStatus",
		"echonetlite_instantaneous_water_heater_bath_water_heating_status",
		"Bath water heater status",
		`{"65":"Heating","66":"Not Heating"}`,
	)
	autoBath := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCAutomaticBathOperation,
		"automaticBathOperation",
		"echonetlite_instantaneous_water_heater_automatic_bath_operation",
		"Bath Auto mode setting",
		`{"65":"Auto ON","66":"Auto OFF"}`,
	)
	supplyTemperature := newPropertyMetricDescriptor(
		echonetlite.EPCTargetSuppliedWaterTemperature,
		"targetSuppliedWaterTemperature", "celsius", gaugeReading,
		"echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius",
		"Set value of hot water temperature",
		deviceLabels,
	)
	bathTemperature := newPropertyMetricDescriptor(
		echonetlite.EPCTargetBathWaterTemperature,
		"targetBathWaterTemperature", "celsius", gaugeReading,
		"echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius",
		"Set value of bath temperature",
		deviceLabels,
	)
	showerSupply := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCShowerHotWaterSupplyStatus,
		"showerHotWaterSupplyStatus",
		"echonetlite_instantaneous_water_heater_shower_hot_water_supply_status",
		"Shower hot water supply status",
		`{"65":"Shower hot water supply ON","66":"Shower hot water supply OFF"}`,
	)
	kitchenSupply := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCKitchenHotWaterSupplyStatus,
		"kitchenHotWaterSupplyStatus",
		"echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status",
		"Kitchen hot water supply status",
		`{"65":"Kitchen hot water supply ON","66":"Kitchen hot water supply OFF"}`,
	)
	bathOperation := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCBathOperationStatus,
		"bathOperationStatus",
		"echonetlite_instantaneous_water_heater_bath_operation_status",
		"Bath operation status monitor",
		`{"65":"Supplying hot water","66":"Stopped","67":"Keeping bath temperature"}`,
	)
	const prefix = "echonetlite_instantaneous_water_heater_"
	hotWaterWarmerSetting := newPropertyMetricDescriptorWithValues(0xD2, "hotWaterWarmerSetting",
		prefix+"hot_water_warmer_setting", "Hot water warmer setting", `{"65":"Hot water warmer operation","66":"Hot water warmer operation resetting"}`)
	remainingAutomaticOperationTime := newPropertyMetricDescriptor(0xDB, "remainingAutomaticOperationTime", "seconds", gaugeReading,
		prefix+"remaining_automatic_operation_time_seconds", "Remaining Automatic operation time", deviceLabels)
	targetBathAdditionalBoilupOperation := newPropertyMetricDescriptorWithValues(0xE4, "targetBathAdditionalBoilupOperation",
		prefix+"target_bath_additional_boilup_operation", "Bath additional boil-up operation setting", `{"65":"Additional boil-up ON","66":"Additional boil-up OFF"}`)
	bathHotWaterAddition := newPropertyMetricDescriptorWithValues(0xE5, "bathHotWaterAddition",
		prefix+"bath_hot_water_addition", "Bath hot water adding operation setting", `{"65":"Hot water addition ON","66":"Hot water addition OFF"}`)
	bathLukewarmWaterFunction := newPropertyMetricDescriptorWithValues(0xE6, "bathLukewarmWaterFunction",
		prefix+"bath_lukewarm_water_function", "Bath water temperature lowering operation setting", `{"65":"Hot water temperature lowering ON","66":"Hot water temperature lowering OFF"}`)
	bathWaterVolume1 := newPropertyMetricDescriptor(0xE7, "bathWaterVolume1", "cubic_meters", gaugeReading,
		prefix+"bath_water_volume1_cubic_meters", "Bath hot water volume setting 1", deviceLabels)
	bathroomPriority := newPropertyMetricDescriptorWithValues(0xE9, "bathroomPriority",
		prefix+"bathroom_priority", "Bathroom priority setting", `{"65":"Bathroom priority ON","66":"Bathroom priority OFF"}`)
	bathWaterVolume3 := newPropertyMetricDescriptor(0xEE, "bathWaterVolume3", "cubic_meters", gaugeReading,
		prefix+"bath_water_volume3_cubic_meters", "Bath hot water volume setting 3", deviceLabels)
	superClass := newSuperClassMetrics(prefix, superClassOperationStatusValues)
	descs := []metricDescriptor{
		hotWaterHeating, bathHeating, autoBath, supplyTemperature, bathTemperature, showerSupply, kitchenSupply, bathOperation,
		hotWaterWarmerSetting,
		remainingAutomaticOperationTime, targetBathAdditionalBoilupOperation, bathHotWaterAddition, bathLukewarmWaterFunction, bathWaterVolume1,
		bathroomPriority, bathWaterVolume3,
	}
	descs = append(descs, superClass.descriptors()...)
	return &InstantaneousWaterHeaterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.InstantaneousWaterHeater) []prometheus.Metric {
				// TODO: Migrate these legacy 0/1 state metrics to ECHONET Lite numeric codes.
				host, eoj := device.Host(), device.EOJ().String()
				metrics := []prometheus.Metric{
					prometheus.MustNewConstMetric(hotWaterHeating.desc, prometheus.GaugeValue, float64(reading.HotWaterHeatingStatus), host, eoj),
					prometheus.MustNewConstMetric(bathHeating.desc, prometheus.GaugeValue, float64(reading.BathWaterHeatingStatus), host, eoj),
				}
				if reading.AutomaticBathOperation != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(autoBath.desc, prometheus.GaugeValue, float64(*reading.AutomaticBathOperation), host, eoj))
				}
				if reading.TargetSuppliedWaterTemperature != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(supplyTemperature.desc, prometheus.GaugeValue, float64(*reading.TargetSuppliedWaterTemperature), host, eoj))
				}
				if reading.TargetBathWaterTemperature != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathTemperature.desc, prometheus.GaugeValue, float64(*reading.TargetBathWaterTemperature), host, eoj))
				}
				if reading.ShowerHotWaterSupplyStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(showerSupply.desc, prometheus.GaugeValue, float64(*reading.ShowerHotWaterSupplyStatus), host, eoj))
				}
				if reading.KitchenHotWaterSupplyStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(kitchenSupply.desc, prometheus.GaugeValue, float64(*reading.KitchenHotWaterSupplyStatus), host, eoj))
				}
				if reading.BathOperationStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathOperation.desc, prometheus.GaugeValue, float64(*reading.BathOperationStatus), host, eoj))
				}

				if value := reading.HotWaterWarmerSetting; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(hotWaterWarmerSetting.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.RemainingAutomaticOperationTime; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(remainingAutomaticOperationTime.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.TargetBathAdditionalBoilupOperation; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(targetBathAdditionalBoilupOperation.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.BathHotWaterAddition; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathHotWaterAddition.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.BathLukewarmWaterFunction; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathLukewarmWaterFunction.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.BathWaterVolume1; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathWaterVolume1.desc, prometheus.GaugeValue, float64(*value)*0.001, host, eoj))
				}
				if value := reading.BathroomPriority; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathroomPriority.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.BathWaterVolume3; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathWaterVolume3.desc, prometheus.GaugeValue, float64(*value)*0.001, host, eoj))
				}
				return superClass.appendSamples(metrics, reading.SuperClass, host, eoj)
			},
		),
	}
}

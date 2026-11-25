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
	hotWaterHeating := newMetricDescriptor(
		echonetlite.EPCHotWaterHeatingStatus,
		"hotWaterHeatingStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_hot_water_heating_status",
		"Hot water heating status",
		deviceLabels,
	)
	bathHeating := newMetricDescriptor(
		echonetlite.EPCBathWaterHeatingStatus,
		"bathWaterHeatingStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_bath_water_heating_status",
		"Bath water heater status",
		deviceLabels,
	)
	autoBath := newMetricDescriptor(
		echonetlite.EPCAutomaticBathOperation,
		"automaticBathOperation", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_automatic_bath_operation",
		"Bath Auto mode setting",
		deviceLabels,
	)
	supplyTemperature := newMetricDescriptor(
		echonetlite.EPCTargetSuppliedWaterTemperature,
		"targetSuppliedWaterTemperature", "celsius", gaugeReading,
		"echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius",
		"Set value of hot water temperature",
		deviceLabels,
	)
	bathTemperature := newMetricDescriptor(
		echonetlite.EPCTargetBathWaterTemperature,
		"targetBathWaterTemperature", "celsius", gaugeReading,
		"echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius",
		"Set value of bath temperature",
		deviceLabels,
	)
	showerSupply := newMetricDescriptor(
		echonetlite.EPCShowerHotWaterSupplyStatus,
		"showerHotWaterSupplyStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_shower_hot_water_supply_status",
		"Shower hot water supply status",
		deviceLabels,
	)
	kitchenSupply := newMetricDescriptor(
		echonetlite.EPCKitchenHotWaterSupplyStatus,
		"kitchenHotWaterSupplyStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status",
		"Kitchen hot water supply status",
		deviceLabels,
	)
	bathOperation := newMetricDescriptor(
		echonetlite.EPCBathOperationStatus,
		"bathOperationStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_bath_operation_status",
		"Bath operation status monitor",
		deviceLabels,
	)
	const prefix = "echonetlite_instantaneous_water_heater_"
	hotWaterWarmerSetting := newMetricDescriptor(0xD2, "hotWaterWarmerSetting", "", gaugeReading,
		prefix+"hot_water_warmer_setting", "Hot water warmer setting", deviceLabels)
	remainingAutomaticOperationTime := newMetricDescriptor(0xDB, "remainingAutomaticOperationTime", "seconds", gaugeReading,
		prefix+"remaining_automatic_operation_time_seconds", "Remaining Automatic operation time", deviceLabels)
	targetBathAdditionalBoilupOperation := newMetricDescriptor(0xE4, "targetBathAdditionalBoilupOperation", "", gaugeReading,
		prefix+"target_bath_additional_boilup_operation", "Bath additional boil-up operation setting", deviceLabels)
	bathHotWaterAddition := newMetricDescriptor(0xE5, "bathHotWaterAddition", "", gaugeReading,
		prefix+"bath_hot_water_addition", "Bath hot water adding operation setting", deviceLabels)
	bathLukewarmWaterFunction := newMetricDescriptor(0xE6, "bathLukewarmWaterFunction", "", gaugeReading,
		prefix+"bath_lukewarm_water_function", "Bath water temperature lowering operation setting", deviceLabels)
	bathWaterVolume1 := newMetricDescriptor(0xE7, "bathWaterVolume1", "cubic_meters", gaugeReading,
		prefix+"bath_water_volume1_cubic_meters", "Bath hot water volume setting 1", deviceLabels)
	bathroomPriority := newMetricDescriptor(0xE9, "bathroomPriority", "", gaugeReading,
		prefix+"bathroom_priority", "Bathroom priority setting", deviceLabels)
	bathWaterVolume3 := newMetricDescriptor(0xEE, "bathWaterVolume3", "cubic_meters", gaugeReading,
		prefix+"bath_water_volume3_cubic_meters", "Bath hot water volume setting 3", deviceLabels)
	superClass := newSuperClassMetrics(prefix)
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
					prometheus.MustNewConstMetric(hotWaterHeating.desc, prometheus.GaugeValue, boolValue(reading.HotWaterHeatingStatus), host, eoj),
					prometheus.MustNewConstMetric(bathHeating.desc, prometheus.GaugeValue, boolValue(reading.BathWaterHeatingStatus), host, eoj),
				}
				if reading.AutomaticBathOperation != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(autoBath.desc, prometheus.GaugeValue, boolValue(*reading.AutomaticBathOperation), host, eoj))
				}
				if reading.TargetSuppliedWaterTemperature != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(supplyTemperature.desc, prometheus.GaugeValue, float64(*reading.TargetSuppliedWaterTemperature), host, eoj))
				}
				if reading.TargetBathWaterTemperature != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathTemperature.desc, prometheus.GaugeValue, float64(*reading.TargetBathWaterTemperature), host, eoj))
				}
				if reading.ShowerHotWaterSupplyStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(showerSupply.desc, prometheus.GaugeValue, boolValue(*reading.ShowerHotWaterSupplyStatus), host, eoj))
				}
				if reading.KitchenHotWaterSupplyStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(kitchenSupply.desc, prometheus.GaugeValue, boolValue(*reading.KitchenHotWaterSupplyStatus), host, eoj))
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

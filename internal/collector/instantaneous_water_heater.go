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
		"hotWaterHeatingStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_hot_water_heating_status",
		"Hot water heating status.",
		deviceLabels,
	)
	bathHeating := newMetricDescriptor(
		"bathWaterHeatingStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_bath_water_heating_status",
		"Bath water heater status.",
		deviceLabels,
	)
	autoBath := newMetricDescriptor(
		"automaticBathOperation", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_automatic_bath_operation",
		"Bath Auto mode setting.",
		deviceLabels,
	)
	supplyTemperature := newMetricDescriptor(
		"targetSuppliedWaterTemperature", "celsius", gaugeReading,
		"echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius",
		"Set value of hot water temperature.",
		deviceLabels,
	)
	bathTemperature := newMetricDescriptor(
		"targetBathWaterTemperature", "celsius", gaugeReading,
		"echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius",
		"Set value of bath temperature.",
		deviceLabels,
	)
	showerSupply := newMetricDescriptor(
		"showerHotWaterSupplyStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_shower_hot_water_supply_status",
		"Shower hot water supply status.",
		deviceLabels,
	)
	kitchenSupply := newMetricDescriptor(
		"kitchenHotWaterSupplyStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status",
		"Kitchen hot water supply status.",
		deviceLabels,
	)
	bathOperation := newMetricDescriptor(
		"bathOperationStatus", "", gaugeReading,
		"echonetlite_instantaneous_water_heater_bath_operation_status",
		"Bath operation status monitor.",
		deviceLabels,
	)
	descs := []metricDescriptor{hotWaterHeating, bathHeating, autoBath, supplyTemperature, bathTemperature, showerSupply, kitchenSupply, bathOperation}
	return &InstantaneousWaterHeaterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.InstantaneousWaterHeater) []prometheus.Metric {
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
				return metrics
			},
		),
	}
}

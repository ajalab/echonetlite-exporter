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
	hotWaterHeating := prometheus.NewDesc(
		"echonetlite_instantaneous_water_heater_hot_water_heating_status",
		"Hot water heating status.",
		deviceLabels, nil,
	)
	bathHeating := prometheus.NewDesc(
		"echonetlite_instantaneous_water_heater_bath_water_heating_status",
		"Bath water heater status.",
		deviceLabels, nil,
	)
	autoBath := prometheus.NewDesc(
		"echonetlite_instantaneous_water_heater_automatic_bath_operation",
		"Bath Auto mode setting.",
		deviceLabels, nil,
	)
	supplyTemperature := prometheus.NewDesc(
		"echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius",
		"Set value of hot water temperature.",
		deviceLabels, nil,
	)
	bathTemperature := prometheus.NewDesc(
		"echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius",
		"Set value of bath temperature.",
		deviceLabels, nil,
	)
	showerSupply := prometheus.NewDesc(
		"echonetlite_instantaneous_water_heater_shower_hot_water_supply_status",
		"Shower hot water supply status.",
		deviceLabels, nil,
	)
	kitchenSupply := prometheus.NewDesc(
		"echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status",
		"Kitchen hot water supply status.",
		deviceLabels, nil,
	)
	bathOperation := prometheus.NewDesc(
		"echonetlite_instantaneous_water_heater_bath_operation_status",
		"Bath operation status monitor.",
		deviceLabels, nil,
	)
	descs := []*prometheus.Desc{hotWaterHeating, bathHeating, autoBath, supplyTemperature, bathTemperature, showerSupply, kitchenSupply, bathOperation}
	return &InstantaneousWaterHeaterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.InstantaneousWaterHeater) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				metrics := []prometheus.Metric{
					prometheus.MustNewConstMetric(hotWaterHeating, prometheus.GaugeValue, boolValue(reading.HotWaterHeatingStatus), host, eoj),
					prometheus.MustNewConstMetric(bathHeating, prometheus.GaugeValue, boolValue(reading.BathWaterHeatingStatus), host, eoj),
				}
				if reading.AutomaticBathOperation != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(autoBath, prometheus.GaugeValue, boolValue(*reading.AutomaticBathOperation), host, eoj))
				}
				if reading.TargetSuppliedWaterTemperature != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(supplyTemperature, prometheus.GaugeValue, float64(*reading.TargetSuppliedWaterTemperature), host, eoj))
				}
				if reading.TargetBathWaterTemperature != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathTemperature, prometheus.GaugeValue, float64(*reading.TargetBathWaterTemperature), host, eoj))
				}
				if reading.ShowerHotWaterSupplyStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(showerSupply, prometheus.GaugeValue, boolValue(*reading.ShowerHotWaterSupplyStatus), host, eoj))
				}
				if reading.KitchenHotWaterSupplyStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(kitchenSupply, prometheus.GaugeValue, boolValue(*reading.KitchenHotWaterSupplyStatus), host, eoj))
				}
				if reading.BathOperationStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(bathOperation, prometheus.GaugeValue, float64(*reading.BathOperationStatus), host, eoj))
				}
				return metrics
			},
		),
	}
}

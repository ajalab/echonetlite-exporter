package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// HybridWaterHeaterCollector collects and publishes metrics for its ECHONET Lite device class.
type HybridWaterHeaterCollector struct {
	*collector
}

// NewHybridWaterHeaterCollector creates a HybridWaterHeaterCollector.
func NewHybridWaterHeaterCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.HybridWaterHeater, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *HybridWaterHeaterCollector {
	solarLinkage := newMetricDescriptor(
		"linkageModeForSolarPowerGeneration", "", gaugeReading,
		"echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation",
		"Linkage mode setting for solar power generation.",
		deviceLabels,
	)
	waterHeating := newMetricDescriptor(
		"waterHeatingStatus", "", gaugeReading,
		"echonetlite_hybrid_water_heater_water_heating_status",
		"Water heating status.",
		deviceLabels,
	)
	heater := newMetricDescriptor(
		"heaterStatus", "", gaugeReading,
		"echonetlite_hybrid_water_heater_heater_status",
		"Heater status.",
		deviceLabels,
	)
	hotWaterSupply := newMetricDescriptor(
		"hotWaterSupplyStatus", "", gaugeReading,
		"echonetlite_hybrid_water_heater_hot_water_supply_status",
		"Hot water supply status.",
		deviceLabels,
	)
	remainingWater := newMetricDescriptor(
		"remainingWater", "cubic_meters", gaugeReading,
		"echonetlite_hybrid_water_heater_remaining_water_cubic_meters",
		"Measured amount of hot water remaining in tank.",
		deviceLabels,
	)
	tankCapacity := newMetricDescriptor(
		"tankCapacity", "cubic_meters", gaugeReading,
		"echonetlite_hybrid_water_heater_tank_capacity_cubic_meters",
		"Tank capacity.",
		deviceLabels,
	)
	descs := []metricDescriptor{solarLinkage, waterHeating, heater, hotWaterSupply, remainingWater, tankCapacity}
	return &HybridWaterHeaterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.HybridWaterHeater) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				metrics := []prometheus.Metric{
					prometheus.MustNewConstMetric(solarLinkage.desc, prometheus.GaugeValue, float64(reading.LinkageModeForSolarPowerGeneration), host, eoj),
				}
				if reading.WaterHeatingStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(waterHeating.desc, prometheus.GaugeValue, boolValue(*reading.WaterHeatingStatus), host, eoj))
				}
				if reading.HeaterStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(heater.desc, prometheus.GaugeValue, boolValue(*reading.HeaterStatus), host, eoj))
				}
				if reading.HotWaterSupplyStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(hotWaterSupply.desc, prometheus.GaugeValue, boolValue(*reading.HotWaterSupplyStatus), host, eoj))
				}
				if reading.RemainingWater != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(remainingWater.desc, prometheus.GaugeValue, float64(*reading.RemainingWater)/1_000, host, eoj))
				}
				if reading.TankCapacity != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(tankCapacity.desc, prometheus.GaugeValue, float64(*reading.TankCapacity)/1_000, host, eoj))
				}
				return metrics
			},
		),
	}
}

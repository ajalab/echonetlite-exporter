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
	solarLinkage := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCLinkageModeForSolarPowerGeneration,
		"linkageModeForSolarPowerGeneration",
		"echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation",
		"Linkage mode setting for solar power generation",
		`{"65":"Mode off","66":"Household consumption","67":"Prioritizing electricity sales","68":"Economic efficiency"}`,
	)
	const waterHeatingValues = `{"65":"Water is heating.","66":"Water is not heating."}`
	waterHeating := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCWaterHeatingStatus,
		"waterHeatingStatus",
		"echonetlite_hybrid_water_heater_water_heating_status",
		"Water heating status",
		waterHeatingValues,
	)
	heater := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCHeaterStatus,
		"heaterStatus",
		"echonetlite_hybrid_water_heater_heater_status",
		"Heater status",
		waterHeatingValues,
	)
	hotWaterSupply := newPropertyMetricDescriptorWithValues(
		echonetlite.EPCHotWaterSupplyStatus,
		"hotWaterSupplyStatus",
		"echonetlite_hybrid_water_heater_hot_water_supply_status",
		"Hot water supply status",
		`{"65":"Supplying hot water","66":"Not supplying hot water"}`,
	)
	remainingWater := newPropertyMetricDescriptor(
		echonetlite.EPCRemainingWater,
		"remainingWater", "cubic_meters", gaugeReading,
		"echonetlite_hybrid_water_heater_remaining_water_cubic_meters",
		"Measured amount of hot water remaining in tank",
		deviceLabels,
	)
	tankCapacity := newPropertyMetricDescriptor(
		echonetlite.EPCTankCapacity,
		"tankCapacity", "cubic_meters", gaugeReading,
		"echonetlite_hybrid_water_heater_tank_capacity_cubic_meters",
		"Tank capacity",
		deviceLabels,
	)
	const prefix = "echonetlite_hybrid_water_heater_"
	automaticWaterHeating := newPropertyMetricDescriptorWithValues(0xB0, "automaticWaterHeating",
		prefix+"automatic_water_heating", "Automatic water heating setting", `{"65":"Automatic water heating","66":"Manual water heating","67":"Water heating manual stop"}`)
	hotWaterSupplyModeForAuxiliaryHeatSourceMachine := newPropertyMetricDescriptorWithValues(0xB6, "hotWaterSupplyModeForAuxiliaryHeatSourceMachine",
		prefix+"hot_water_supply_mode_for_auxiliary_heat_source_machine", "Hot water supply mode setting for auxiliary heat source machine", `{"65":"Set","66":"No setting"}`)
	heaterModeForAuxiliaryHeatSourceMachine := newPropertyMetricDescriptorWithValues(0xB7, "heaterModeForAuxiliaryHeatSourceMachine",
		prefix+"heater_mode_for_auxiliary_heat_source_machine", "Heater mode setting for auxiliary heat source machine.", `{"65":"Set","66":"No setting"}`)
	superClass := newSuperClassMetrics(prefix, superClassOperationStatusValues)
	descs := []metricDescriptor{
		solarLinkage, waterHeating, heater, hotWaterSupply, remainingWater, tankCapacity,
		automaticWaterHeating,
		hotWaterSupplyModeForAuxiliaryHeatSourceMachine, heaterModeForAuxiliaryHeatSourceMachine,
	}
	descs = append(descs, superClass.descriptors()...)
	return &HybridWaterHeaterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.HybridWaterHeater) []prometheus.Metric {
				// TODO: Migrate these legacy 0/1 state metrics to ECHONET Lite numeric codes.
				host, eoj := device.Host(), device.EOJ().String()
				metrics := []prometheus.Metric{
					prometheus.MustNewConstMetric(solarLinkage.desc, prometheus.GaugeValue, float64(reading.LinkageModeForSolarPowerGeneration), host, eoj),
				}
				if reading.WaterHeatingStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(waterHeating.desc, prometheus.GaugeValue, float64(*reading.WaterHeatingStatus), host, eoj))
				}
				if reading.HeaterStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(heater.desc, prometheus.GaugeValue, float64(*reading.HeaterStatus), host, eoj))
				}
				if reading.HotWaterSupplyStatus != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(hotWaterSupply.desc, prometheus.GaugeValue, float64(*reading.HotWaterSupplyStatus), host, eoj))
				}
				if reading.RemainingWater != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(remainingWater.desc, prometheus.GaugeValue, float64(*reading.RemainingWater)/1_000, host, eoj))
				}
				if reading.TankCapacity != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(tankCapacity.desc, prometheus.GaugeValue, float64(*reading.TankCapacity)/1_000, host, eoj))
				}

				if value := reading.AutomaticWaterHeating; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(automaticWaterHeating.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.HotWaterSupplyModeForAuxiliaryHeatSourceMachine; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(hotWaterSupplyModeForAuxiliaryHeatSourceMachine.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.HeaterModeForAuxiliaryHeatSourceMachine; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(heaterModeForAuxiliaryHeatSourceMachine.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				return superClass.appendSamples(metrics, reading.SuperClass, host, eoj)
			},
		),
	}
}

package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// StorageBatteryCollector collects and publishes metrics for its ECHONET Lite device class.
type StorageBatteryCollector struct {
	*collector
}

// NewStorageBatteryCollector creates a StorageBatteryCollector.
func NewStorageBatteryCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.StorageBattery, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *StorageBatteryCollector {
	chargeableEnergy := newMetricDescriptor(
		"acChargeableElectricEnergy", "joules", gaugeReading,
		"echonetlite_storage_battery_ac_chargeable_electric_energy_joules",
		"AC chargeable electric energy.",
		deviceLabels,
	)
	dischargeableEnergy := newMetricDescriptor(
		"acDischargeableElectricEnergy", "joules", gaugeReading,
		"echonetlite_storage_battery_ac_dischargeable_electric_energy_joules",
		"AC dischargeable electric energy.",
		deviceLabels,
	)
	chargingEnergy := newMetricDescriptor(
		"acCumulativeChargingElectricEnergy", "joules", cumulativeReading,
		"echonetlite_storage_battery_ac_charging_electric_energy_joules_total",
		"AC measured cumulative charging electric energy.",
		deviceLabels,
	)
	dischargingEnergy := newMetricDescriptor(
		"acCumulativeDischargingElectricEnergy", "joules", cumulativeReading,
		"echonetlite_storage_battery_ac_discharging_electric_energy_joules_total",
		"AC measured cumulative discharging electric energy.",
		deviceLabels,
	)
	descs := []metricDescriptor{chargeableEnergy, dischargeableEnergy, chargingEnergy, dischargingEnergy}
	return &StorageBatteryCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.StorageBattery) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				return []prometheus.Metric{
					prometheus.MustNewConstMetric(chargeableEnergy.desc, prometheus.GaugeValue, float64(reading.ACChargeableElectricEnergy)*joulesPerWattHour, host, eoj),
					prometheus.MustNewConstMetric(dischargeableEnergy.desc, prometheus.GaugeValue, float64(reading.ACDischargeableElectricEnergy)*joulesPerWattHour, host, eoj),
					prometheus.MustNewConstMetric(chargingEnergy.desc, prometheus.CounterValue, float64(reading.ACCumulativeChargingElectricEnergy)*joulesPerWattHour, host, eoj),
					prometheus.MustNewConstMetric(dischargingEnergy.desc, prometheus.CounterValue, float64(reading.ACCumulativeDischargingElectricEnergy)*joulesPerWattHour, host, eoj),
				}
			},
		),
	}
}

package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// PVPowerGenerationCollector collects and publishes metrics for its ECHONET Lite device class.
type PVPowerGenerationCollector struct {
	*collector
}

// NewPVPowerGenerationCollector creates a PVPowerGenerationCollector.
func NewPVPowerGenerationCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.PVPowerGeneration, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *PVPowerGenerationCollector {
	power := newMetricDescriptor(
		echonetlite.EPCInstantaneousElectricPowerGeneration,
		"instantaneousElectricPowerGeneration", "watts", instantaneousReading,
		"echonetlite_pv_power_generation_electric_power_generation_watts",
		"Measured instantaneous amount of electricity generated",
		deviceLabels,
	)
	energy := newMetricDescriptor(
		echonetlite.EPCCumulativeElectricEnergyOfGeneration,
		"cumulativeElectricEnergyOfGeneration", "joules", cumulativeReading,
		"echonetlite_pv_power_generation_electric_energy_of_generation_joules_total",
		"Measured cumulative amount of electric energy generated",
		deviceLabels,
	)
	descs := []metricDescriptor{power, energy}
	return &PVPowerGenerationCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.PVPowerGeneration) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				return []prometheus.Metric{
					prometheus.MustNewConstMetric(power.desc, prometheus.GaugeValue, float64(reading.InstantaneousElectricPowerGeneration), host, eoj),
					prometheus.MustNewConstMetric(energy.desc, prometheus.CounterValue, float64(reading.CumulativeElectricEnergyOfGeneration)*joulesPerWattHour, host, eoj),
				}
			},
		),
	}
}

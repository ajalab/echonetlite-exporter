package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// MultipleInputPCSCollector collects and publishes metrics for its ECHONET Lite device class.
type MultipleInputPCSCollector struct {
	*collector
}

// NewMultipleInputPCSCollector creates a MultipleInputPCSCollector.
func NewMultipleInputPCSCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.MultipleInputPCS, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *MultipleInputPCSCollector {
	power := prometheus.NewDesc(
		"echonetlite_multiple_input_pcs_electric_power_watts",
		"Measured instantaneous electric power.",
		deviceLabels, nil,
	)
	normalEnergy := prometheus.NewDesc(
		"echonetlite_multiple_input_pcs_normal_direction_electric_energy_joules_total",
		"Measured cumulative amount of electric energy (normal direction).",
		deviceLabels, nil,
	)
	reverseEnergy := prometheus.NewDesc(
		"echonetlite_multiple_input_pcs_reverse_direction_electric_energy_joules_total",
		"Measured cumulative amount of electric energy (reverse direction).",
		deviceLabels, nil,
	)
	descs := []*prometheus.Desc{power, normalEnergy, reverseEnergy}
	return &MultipleInputPCSCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.MultipleInputPCS) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				return []prometheus.Metric{
					prometheus.MustNewConstMetric(power, prometheus.GaugeValue, float64(reading.InstantaneousElectricPower), host, eoj),
					prometheus.MustNewConstMetric(normalEnergy, prometheus.CounterValue, float64(reading.NormalDirectionElectricEnergy)*joulesPerWattHour, host, eoj),
					prometheus.MustNewConstMetric(reverseEnergy, prometheus.CounterValue, float64(reading.ReverseDirectionElectricEnergy)*joulesPerWattHour, host, eoj),
				}
			},
		),
	}
}

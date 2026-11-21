package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// WattHourMeterCollector collects and publishes metrics for its ECHONET Lite device class.
type WattHourMeterCollector struct {
	*collector
}

// NewWattHourMeterCollector creates a WattHourMeterCollector.
func NewWattHourMeterCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.WattHourMeter, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *WattHourMeterCollector {
	energy := prometheus.NewDesc(
		"echonetlite_watt_hour_meter_electric_energy_joules_total",
		"Cumulative amounts of electric energy measurement value.",
		deviceLabels, nil,
	)
	descs := []*prometheus.Desc{energy}
	return &WattHourMeterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.WattHourMeter) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				return []prometheus.Metric{
					prometheus.MustNewConstMetric(energy, prometheus.CounterValue, float64(reading.CumulativeElectricEnergy)*reading.CumulativeAmountsOfElectricEnergyUnit*joulesPerKilowattHour, host, eoj),
				}
			},
		),
	}
}

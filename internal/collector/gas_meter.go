package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// GasMeterCollector collects and publishes metrics for its ECHONET Lite device class.
type GasMeterCollector struct {
	*collector
}

// NewGasMeterCollector creates a GasMeterCollector.
func NewGasMeterCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.GasMeter, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *GasMeterCollector {
	consumption := newMetricDescriptor(
		echonetlite.EPCCumulativeGasConsumption,
		"cumulativeGasConsumption", "cubic_meters", cumulativeReading,
		"echonetlite_gas_meter_gas_consumption_cubic_meters_total",
		"Cumulative amount of gas consumption measurement value",
		deviceLabels,
	)
	descs := []metricDescriptor{consumption}
	return &GasMeterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.GasMeter) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				return []prometheus.Metric{
					prometheus.MustNewConstMetric(consumption.desc, prometheus.CounterValue, float64(reading.CumulativeGasConsumption)/1_000, host, eoj),
				}
			},
		),
	}
}

package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// WaterFlowMeterCollector collects and publishes metrics for its ECHONET Lite device class.
type WaterFlowMeterCollector struct {
	*collector
}

// NewWaterFlowMeterCollector creates a WaterFlowMeterCollector.
func NewWaterFlowMeterCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.WaterFlowMeter, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *WaterFlowMeterCollector {
	flow := prometheus.NewDesc(
		"echonetlite_water_flow_meter_flowing_water_cubic_meters_total",
		"Measured cumulative amount of flowing water.",
		deviceLabels, nil,
	)
	descs := []*prometheus.Desc{flow}
	return &WaterFlowMeterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.WaterFlowMeter) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				return []prometheus.Metric{
					prometheus.MustNewConstMetric(flow, prometheus.CounterValue, float64(reading.CumulativeFlowingWater)*reading.UnitForCumulativeFlowingWater, host, eoj),
				}
			},
		),
	}
}

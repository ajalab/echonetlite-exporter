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
	flow := newPropertyMetricDescriptor(
		echonetlite.EPCCumulativeFlowingWater,
		"cumulativeFlowingWater", "cubic_meters", cumulativeReading,
		"echonetlite_water_flow_meter_flowing_water_cubic_meters_total",
		"Measured cumulative amount of flowing water",
		deviceLabels,
	)
	const prefix = "echonetlite_water_flow_meter_"
	abnormalDetection := newPropertyMetricDescriptorWithValues(0xE3, "abnormalDetection",
		prefix+"abnormal_detection", "Detection of abnormal value in metering data", `{"65":"Detected","66":"Not detected"}`)
	securityDataInformation := newPropertyMetricDescriptor(0xE4, "securityDataInformation", "", gaugeReading,
		prefix+"security_data_information", "Security data information", deviceLabels)
	superClass := newSuperClassMetrics(prefix, superClassOperationStatusValues)
	descs := []metricDescriptor{
		flow,
		abnormalDetection,
		securityDataInformation,
	}
	descs = append(descs, superClass.descriptors()...)
	return &WaterFlowMeterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.WaterFlowMeter) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				metrics := []prometheus.Metric{
					prometheus.MustNewConstMetric(flow.desc, prometheus.CounterValue, float64(reading.CumulativeFlowingWater)*reading.UnitForCumulativeFlowingWater, host, eoj),
				}

				if value := reading.AbnormalDetection; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(abnormalDetection.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.SecurityDataInformation; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(securityDataInformation.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				return superClass.appendSamples(metrics, reading.SuperClass, host, eoj)
			},
		),
	}
}

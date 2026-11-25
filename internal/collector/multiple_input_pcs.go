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
	power := newPropertyMetricDescriptor(
		echonetlite.EPCMultipleInputPCSInstantaneousElectricPower,
		"instantaneousElectricPower", "watts", instantaneousReading,
		"echonetlite_multiple_input_pcs_electric_power_watts",
		"Measured instantaneous electric power",
		deviceLabels,
	)
	normalEnergy := newPropertyMetricDescriptor(
		echonetlite.EPCMultipleInputPCSNormalDirectionElectricEnergy,
		"normalDirectionElectricEnergy", "joules", cumulativeReading,
		"echonetlite_multiple_input_pcs_normal_direction_electric_energy_joules_total",
		"Measured cumulative amount of electric energy (normal direction)",
		deviceLabels,
	)
	reverseEnergy := newPropertyMetricDescriptor(
		echonetlite.EPCMultipleInputPCSReverseDirectionElectricEnergy,
		"reverseDirectionElectricEnergy", "joules", cumulativeReading,
		"echonetlite_multiple_input_pcs_reverse_direction_electric_energy_joules_total",
		"Measured cumulative amount of electric energy (reverse direction)",
		deviceLabels,
	)
	const prefix = "echonetlite_multiple_input_pcs_"
	powerSystemInterconnectionStatus := newPropertyMetricDescriptorWithValues(0xD0, "powerSystemInterconnectionStatus",
		prefix+"power_system_interconnection_status", "System interconnection status", `{"0":"System-interconnected type (reverse power flow acceptable)","1":"Independent type","2":"System-interconnected type (reverse power flow not acceptable)"}`)
	superClass := newSuperClassMetrics(prefix, `{"48":"On","49":"Off"}`)
	descs := []metricDescriptor{
		power, normalEnergy, reverseEnergy,
		powerSystemInterconnectionStatus,
	}
	descs = append(descs, superClass.descriptors()...)
	return &MultipleInputPCSCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.MultipleInputPCS) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				metrics := []prometheus.Metric{
					prometheus.MustNewConstMetric(power.desc, prometheus.GaugeValue, float64(reading.InstantaneousElectricPower), host, eoj),
					prometheus.MustNewConstMetric(normalEnergy.desc, prometheus.CounterValue, float64(reading.NormalDirectionElectricEnergy)*joulesPerWattHour, host, eoj),
					prometheus.MustNewConstMetric(reverseEnergy.desc, prometheus.CounterValue, float64(reading.ReverseDirectionElectricEnergy)*joulesPerWattHour, host, eoj),
				}

				if value := reading.PowerSystemInterconnectionStatus; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(powerSystemInterconnectionStatus.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				return superClass.appendSamples(metrics, reading.SuperClass, host, eoj)
			},
		),
	}
}

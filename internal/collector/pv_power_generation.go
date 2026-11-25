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
	power := newPropertyMetricDescriptor(
		echonetlite.EPCInstantaneousElectricPowerGeneration,
		"instantaneousElectricPowerGeneration", "watts", instantaneousReading,
		"echonetlite_pv_power_generation_electric_power_generation_watts",
		"Measured instantaneous amount of electricity generated",
		deviceLabels,
	)
	energy := newPropertyMetricDescriptor(
		echonetlite.EPCCumulativeElectricEnergyOfGeneration,
		"cumulativeElectricEnergyOfGeneration", "joules", cumulativeReading,
		"echonetlite_pv_power_generation_electric_energy_of_generation_joules_total",
		"Measured cumulative amount of electric energy generated",
		deviceLabels,
	)
	const prefix = "echonetlite_pv_power_generation_"
	outputPowerControl1 := newPropertyMetricDescriptor(0xA0, "outputPowerControl1", "ratio", gaugeReading,
		prefix+"output_power_control1_ratio", "Output power control setting 1", deviceLabels)
	outputPowerControl2 := newPropertyMetricDescriptor(0xA1, "outputPowerControl2", "watts", gaugeReading,
		prefix+"output_power_control2_watts", "Output power control setting 2", deviceLabels)
	upperLimitClip := newPropertyMetricDescriptor(0xB4, "upperLimitClip", "watts", gaugeReading,
		prefix+"upper_limit_clip_watts", "Upper limit clip setting value", deviceLabels)
	approvedCapacity := newPropertyMetricDescriptor(0xC3, "approvedCapacity", "watts", gaugeReading,
		prefix+"approved_capacity_watts", "Capacity approved by equipment", deviceLabels)
	powerSystemInterconnectionStatus := newPropertyMetricDescriptorWithValues(0xD0, "powerSystemInterconnectionStatus",
		prefix+"power_system_interconnection_status", "System-interconnected type", `{"0":"System-interconnected type (reverse power flow acceptable)","1":"Independent type","2":"System-interconnected type (reverse power flow not acceptable)","3":"Unknown"}`)
	outputPowerRestraintStatus := newPropertyMetricDescriptorWithValues(0xD1, "outputPowerRestraintStatus",
		prefix+"output_power_restraint_status", "Output power restraint status", `{"65":"Ongoing restraint (output power control)","66":"Ongoing restraint (except output power control)","67":"Ongoing restraint (reason for restraint is unknown)","68":"Not restraining","69":"Unknown"}`)
	cumulativeElectricEnergySold := newPropertyMetricDescriptor(0xE3, "cumulativeElectricEnergySold", "joules", cumulativeReading,
		prefix+"electric_energy_sold_joules_total", "Measured cumulative amount of electric energy sold", deviceLabels)
	powerGenerationOutputLimit1 := newPropertyMetricDescriptor(0xE5, "powerGenerationOutputLimit1", "ratio", gaugeReading,
		prefix+"power_generation_output_limit1_ratio", "Power generation output limit setting 1", deviceLabels)
	powerGenerationOutputLimit2 := newPropertyMetricDescriptor(0xE6, "powerGenerationOutputLimit2", "watts", gaugeReading,
		prefix+"power_generation_output_limit2_watts", "Power generation output limit setting 2", deviceLabels)
	limitElectricEnergySold := newPropertyMetricDescriptor(0xE7, "limitElectricEnergySold", "watts", gaugeReading,
		prefix+"limit_electric_energy_sold_watts", "Limit setting for the amount of electricity sold", deviceLabels)
	ratedElectricPowerOfgeneration := newPropertyMetricDescriptor(0xE8, "ratedElectricPowerOfgeneration", "watts", gaugeReading,
		prefix+"rated_electric_power_ofgeneration_watts", "Rated power generation output (System-interconnected)", deviceLabels)
	ratedElectricPowerOfgenerationIndependent := newPropertyMetricDescriptor(0xE9, "ratedElectricPowerOfgenerationIndependent", "watts", gaugeReading,
		prefix+"rated_electric_power_ofgeneration_independent_watts", "Rated power generation output (Independent)", deviceLabels)
	superClass := newSuperClassMetrics(prefix, superClassOperationStatusValues)
	descs := []metricDescriptor{
		power, energy,
		outputPowerControl1,
		outputPowerControl2, upperLimitClip, approvedCapacity, powerSystemInterconnectionStatus, outputPowerRestraintStatus,
		cumulativeElectricEnergySold, powerGenerationOutputLimit1, powerGenerationOutputLimit2, limitElectricEnergySold, ratedElectricPowerOfgeneration,
		ratedElectricPowerOfgenerationIndependent,
	}
	descs = append(descs, superClass.descriptors()...)
	return &PVPowerGenerationCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.PVPowerGeneration) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				metrics := []prometheus.Metric{
					prometheus.MustNewConstMetric(power.desc, prometheus.GaugeValue, float64(reading.InstantaneousElectricPowerGeneration), host, eoj),
					prometheus.MustNewConstMetric(energy.desc, prometheus.CounterValue, float64(reading.CumulativeElectricEnergyOfGeneration)*joulesPerWattHour, host, eoj),
				}

				if value := reading.OutputPowerControl1; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(outputPowerControl1.desc, prometheus.GaugeValue, float64(*value)*0.01, host, eoj))
				}
				if value := reading.OutputPowerControl2; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(outputPowerControl2.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.UpperLimitClip; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(upperLimitClip.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.ApprovedCapacity; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(approvedCapacity.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.PowerSystemInterconnectionStatus; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(powerSystemInterconnectionStatus.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.OutputPowerRestraintStatus; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(outputPowerRestraintStatus.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.CumulativeElectricEnergySold; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(cumulativeElectricEnergySold.desc, prometheus.CounterValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.PowerGenerationOutputLimit1; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(powerGenerationOutputLimit1.desc, prometheus.GaugeValue, float64(*value)*0.01, host, eoj))
				}
				if value := reading.PowerGenerationOutputLimit2; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(powerGenerationOutputLimit2.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.LimitElectricEnergySold; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(limitElectricEnergySold.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.RatedElectricPowerOfgeneration; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(ratedElectricPowerOfgeneration.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.RatedElectricPowerOfgenerationIndependent; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(ratedElectricPowerOfgenerationIndependent.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				return superClass.appendSamples(metrics, reading.SuperClass, host, eoj)
			},
		),
	}
}

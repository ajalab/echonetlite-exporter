package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// DRElectricEnergyMeterCollector collects and publishes metrics for its ECHONET Lite device class.
type DRElectricEnergyMeterCollector struct {
	*collector
}

// NewDRElectricEnergyMeterCollector creates a DRElectricEnergyMeterCollector.
func NewDRElectricEnergyMeterCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.DRElectricEnergyMeter, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *DRElectricEnergyMeterCollector {
	acPower := newMetricDescriptor(
		"acInstantaneousElectricPower", "watts", instantaneousReading,
		"echonetlite_dr_electric_energy_meter_ac_electric_power_watts",
		"Measured instantaneous electric power (AC input/output).",
		deviceLabels,
	)
	independentPower := newMetricDescriptor(
		"independentOperationInstantaneousElectricPower", "watts", instantaneousReading,
		"echonetlite_dr_electric_energy_meter_independent_operation_electric_power_watts",
		"Measured instantaneous electric power (output during a power outage).",
		deviceLabels,
	)
	inputEnergy := newMetricDescriptor(
		"acInputCumulativeElectricEnergy", "joules", cumulativeReading,
		"echonetlite_dr_electric_energy_meter_ac_input_electric_energy_joules_total",
		"Measured cumulative amounts of electric energy (AC input).",
		deviceLabels,
	)
	outputEnergy := newMetricDescriptor(
		"acOutputCumulativeElectricEnergy", "joules", cumulativeReading,
		"echonetlite_dr_electric_energy_meter_ac_output_electric_energy_joules_total",
		"Measured cumulative amounts of electric energy (AC output).",
		deviceLabels,
	)
	independentEnergy := newMetricDescriptor(
		"independentOperationCumulativeElectricEnergy", "joules", cumulativeReading,
		"echonetlite_dr_electric_energy_meter_independent_operation_electric_energy_joules_total",
		"Measured cumulative amounts of electric energy (output during a power outage).",
		deviceLabels,
	)
	descs := []metricDescriptor{acPower, independentPower, inputEnergy, outputEnergy, independentEnergy}
	return &DRElectricEnergyMeterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.DRElectricEnergyMeter) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				scale := reading.CumulativeAmountsOfElectricEnergyUnit * joulesPerKilowattHour
				var metrics []prometheus.Metric
				if reading.ACInstantaneousElectricPower != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acPower.desc, prometheus.GaugeValue, float64(*reading.ACInstantaneousElectricPower), host, eoj))
				}
				if reading.IndependentOperationInstantaneousElectricPower != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(independentPower.desc, prometheus.GaugeValue, float64(*reading.IndependentOperationInstantaneousElectricPower), host, eoj))
				}
				if reading.ACInputCumulativeElectricEnergy != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(inputEnergy.desc, prometheus.CounterValue, float64(*reading.ACInputCumulativeElectricEnergy)*scale, host, eoj))
				}
				if reading.ACOutputCumulativeElectricEnergy != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(outputEnergy.desc, prometheus.CounterValue, float64(*reading.ACOutputCumulativeElectricEnergy)*scale, host, eoj))
				}
				if reading.IndependentOperationCumulativeElectricEnergy != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(independentEnergy.desc, prometheus.CounterValue, float64(*reading.IndependentOperationCumulativeElectricEnergy)*scale, host, eoj))
				}
				return metrics
			},
		),
	}
}

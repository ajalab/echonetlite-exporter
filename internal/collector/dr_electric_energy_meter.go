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
	acPower := prometheus.NewDesc(
		"echonetlite_dr_electric_energy_meter_ac_inout_electric_power_watts",
		"Measured instantaneous electric power (AC input/output).",
		deviceLabels, nil,
	)
	independentPower := prometheus.NewDesc(
		"echonetlite_dr_electric_energy_meter_independent_operation_electric_power_watts",
		"Measured instantaneous electric power (output during a power outage).",
		deviceLabels, nil,
	)
	inputEnergy := prometheus.NewDesc(
		"echonetlite_dr_electric_energy_meter_ac_input_electric_energy_joules_total",
		"Measured cumulative amounts of electric energy (AC input).",
		deviceLabels, nil,
	)
	outputEnergy := prometheus.NewDesc(
		"echonetlite_dr_electric_energy_meter_ac_output_electric_energy_joules_total",
		"Measured cumulative amounts of electric energy (AC output).",
		deviceLabels, nil,
	)
	independentEnergy := prometheus.NewDesc(
		"echonetlite_dr_electric_energy_meter_independent_operation_electric_energy_joules_total",
		"Measured cumulative amounts of electric energy (output during a power outage).",
		deviceLabels, nil,
	)
	descs := []*prometheus.Desc{acPower, independentPower, inputEnergy, outputEnergy, independentEnergy}
	return &DRElectricEnergyMeterCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.DRElectricEnergyMeter) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				scale := reading.CumulativeAmountsOfElectricEnergyUnit * joulesPerKilowattHour
				var metrics []prometheus.Metric
				if reading.ACInstantaneousElectricPower != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acPower, prometheus.GaugeValue, float64(*reading.ACInstantaneousElectricPower), host, eoj))
				}
				if reading.IndependentOperationInstantaneousElectricPower != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(independentPower, prometheus.GaugeValue, float64(*reading.IndependentOperationInstantaneousElectricPower), host, eoj))
				}
				if reading.ACInputCumulativeElectricEnergy != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(inputEnergy, prometheus.CounterValue, float64(*reading.ACInputCumulativeElectricEnergy)*scale, host, eoj))
				}
				if reading.ACOutputCumulativeElectricEnergy != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(outputEnergy, prometheus.CounterValue, float64(*reading.ACOutputCumulativeElectricEnergy)*scale, host, eoj))
				}
				if reading.IndependentOperationCumulativeElectricEnergy != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(independentEnergy, prometheus.CounterValue, float64(*reading.IndependentOperationCumulativeElectricEnergy)*scale, host, eoj))
				}
				return metrics
			},
		),
	}
}

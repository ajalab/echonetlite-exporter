package collector

import (
	"context"
	"strconv"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// PowerDistributionBoardMeteringCollector collects and publishes metrics for its ECHONET Lite device class.
type PowerDistributionBoardMeteringCollector struct {
	*collector
}

// NewPowerDistributionBoardMeteringCollector creates a PowerDistributionBoardMeteringCollector.
func NewPowerDistributionBoardMeteringCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.PowerDistributionBoardMetering, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *PowerDistributionBoardMeteringCollector {
	labels := []string{"host", "eoj", "channel"}
	power := newMetricDescriptor(
		"instantaneousElectricPowerListSimplex", "watts", instantaneousReading,
		"echonetlite_power_distribution_board_metering_electric_power_list_simplex_watts",
		"Measured instantaneous power consumption list (simplex).",
		labels,
	)
	energy := newMetricDescriptor(
		"cumulativeElectricEnergyListSimplex", "joules", cumulativeReading,
		"echonetlite_power_distribution_board_metering_electric_energy_list_simplex_joules_total",
		"Measured cumulative amount of electric power consumption list (simplex).",
		labels,
	)
	return &PowerDistributionBoardMeteringCollector{
		collector: newCollector(interval, targets, []metricDescriptor{power, energy}, get,
			func(device echonetlite.Device, reading *echonetlite.PowerDistributionBoardMetering) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				var metrics []prometheus.Metric
				powers := reading.InstantaneousElectricPowerListSimplex
				for i, value := range powers.InstantaneousElectricPower {
					if value != nil {
						channel := strconv.Itoa(int(powers.StartChannel) + i)
						metrics = append(metrics, prometheus.MustNewConstMetric(power.desc, prometheus.GaugeValue, float64(*value), host, eoj, channel))
					}
				}
				energies := reading.CumulativeElectricEnergyListSimplex
				scale := reading.UnitForCumulativeElectricEnergy * joulesPerKilowattHour
				for i, value := range energies.ElectricEnergy {
					if value != nil {
						channel := strconv.Itoa(int(energies.StartChannel) + i)
						metrics = append(metrics, prometheus.MustNewConstMetric(energy.desc, prometheus.CounterValue, float64(*value)*scale, host, eoj, channel))
					}
				}
				return metrics
			},
		),
	}
}

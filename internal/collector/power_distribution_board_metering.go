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
		echonetlite.EPCInstantaneousElectricPowerListSimplex,
		"instantaneousElectricPowerListSimplex", "watts", instantaneousReading,
		"echonetlite_power_distribution_board_metering_electric_power_list_simplex_watts",
		"Measured instantaneous power consumption list (simplex)",
		labels,
	)
	energy := newMetricDescriptor(
		echonetlite.EPCCumulativeElectricEnergyListSimplex,
		"cumulativeElectricEnergyListSimplex", "joules", cumulativeReading,
		"echonetlite_power_distribution_board_metering_electric_energy_list_simplex_joules_total",
		"Measured cumulative amount of electric power consumption list (simplex)",
		labels,
	)
	const prefix = "echonetlite_power_distribution_board_metering_"
	masterRatedCapacity := newMetricDescriptor(0xB0, "masterRatedCapacity", "amperes", gaugeReading,
		prefix+"master_rated_capacity_amperes", "Master rated capacity", deviceLabels)
	numberOfMeasurementChannelsSimplex := newMetricDescriptor(0xB1, "numberOfMeasurementChannelsSimplex", "", gaugeReading,
		prefix+"number_of_measurement_channels_simplex", "Number of measurement channels (simplex)", deviceLabels)
	numberOfMeasurementChannelsDuplex := newMetricDescriptor(0xB8, "numberOfMeasurementChannelsDuplex", "", gaugeReading,
		prefix+"number_of_measurement_channels_duplex", "Number of measurement channels (duplex)", deviceLabels)
	normalDirectionCumulativeElectricEnergy := newMetricDescriptor(0xC0, "normalDirectionCumulativeElectricEnergy", "joules", cumulativeReading,
		prefix+"normal_direction_electric_energy_joules_total", "Measured cumulative amount of electric energy (normal direction)", deviceLabels)
	reverseDirectionCumulativeElectricEnergy := newMetricDescriptor(0xC1, "reverseDirectionCumulativeElectricEnergy", "joules", cumulativeReading,
		prefix+"reverse_direction_electric_energy_joules_total", "Measured cumulative amount of electric energy (reverse direction)", deviceLabels)
	instantaneousElectricPower := newMetricDescriptor(0xC6, "instantaneousElectricPower", "watts", instantaneousReading,
		prefix+"electric_power_watts", "Measured instantaneous amount of electric energy", deviceLabels)
	phaseLabels := []string{"host", "eoj", "phase"}
	channelPhaseLabels := []string{"host", "eoj", "channel", "phase"}
	directionLabels := []string{"host", "eoj", "channel", "direction"}
	voltageLabels := []string{"host", "eoj", "line_pair"}
	currentSimplex := newMetricDescriptor(0xB5, "instantaneousCurrentListSimplex", "amperes", instantaneousReading,
		prefix+"current_list_simplex_amperes", "Measured instantaneous current list (simplex)", channelPhaseLabels)
	energyDuplex := newMetricDescriptor(0xBA, "cumulativeElectricEnergyListDuplex", "joules", cumulativeReading,
		prefix+"electric_energy_list_duplex_joules_total", "Measured cumulative amount of electric power consumption list (duplex)", directionLabels)
	currentDuplex := newMetricDescriptor(0xBC, "instantaneousCurrentListDuplex", "amperes", instantaneousReading,
		prefix+"current_list_duplex_amperes", "Measured instantaneous current list (duplex)", channelPhaseLabels)
	powerDuplex := newMetricDescriptor(0xBE, "instantaneousElectricPowerListDuplex", "watts", instantaneousReading,
		prefix+"electric_power_list_duplex_watts", "Measured instantaneous power consumption list (duplex)", labels)
	current := newMetricDescriptor(0xC7, "instantaneousCurrent", "amperes", instantaneousReading,
		prefix+"current_amperes", "Measured instantaneous currents", phaseLabels)
	voltage := newMetricDescriptor(0xC8, "instantaneousVoltage", "volts", instantaneousReading,
		prefix+"voltage_volts", "Measured instantaneous voltages", voltageLabels)
	superClass := newSuperClassMetrics(prefix)
	descs := []metricDescriptor{
		power, energy,
		masterRatedCapacity,
		numberOfMeasurementChannelsSimplex, numberOfMeasurementChannelsDuplex, normalDirectionCumulativeElectricEnergy, reverseDirectionCumulativeElectricEnergy, instantaneousElectricPower,
		currentSimplex, energyDuplex, currentDuplex, powerDuplex, current,
		voltage,
	}
	descs = append(descs, superClass.descriptors()...)
	var channelEnergy, channelCurrent [32]metricDescriptor
	for i := range channelEnergy {
		channel := strconv.Itoa(i + 1)
		epc := 0xD0 + uint8(i)
		shortName := "measurementChannel" + channel
		help := "Measurement channel " + channel
		channelEnergy[i] = newMetricDescriptor(epc, shortName, "joules", cumulativeReading,
			prefix+"measurement_channel"+channel+"_joules_total", help, deviceLabels)
		channelCurrent[i] = newMetricDescriptor(epc, shortName, "amperes", instantaneousReading,
			prefix+"measurement_channel"+channel+"_amperes", help, phaseLabels)
		descs = append(descs, channelEnergy[i], channelCurrent[i])
	}
	return &PowerDistributionBoardMeteringCollector{
		collector: newCollector(interval, targets, descs, get,
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

				if value := reading.MasterRatedCapacity; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(masterRatedCapacity.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.NumberOfMeasurementChannelsSimplex; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(numberOfMeasurementChannelsSimplex.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.NumberOfMeasurementChannelsDuplex; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(numberOfMeasurementChannelsDuplex.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.NormalDirectionCumulativeElectricEnergy; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(normalDirectionCumulativeElectricEnergy.desc, prometheus.CounterValue, float64(*value)*scale, host, eoj))
				}
				if value := reading.ReverseDirectionCumulativeElectricEnergy; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(reverseDirectionCumulativeElectricEnergy.desc, prometheus.CounterValue, float64(*value)*scale, host, eoj))
				}
				if value := reading.InstantaneousElectricPower; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(instantaneousElectricPower.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if list := reading.InstantaneousCurrentListSimplex; list != nil {
					for i, value := range list.InstantaneousCurrent {
						channel := strconv.Itoa(int(list.StartChannel) + i)
						if value.RPhase != nil {
							metrics = append(metrics, prometheus.MustNewConstMetric(currentSimplex.desc, prometheus.GaugeValue, float64(*value.RPhase)*0.1, host, eoj, channel, "r"))
						}
						if value.TPhase != nil {
							metrics = append(metrics, prometheus.MustNewConstMetric(currentSimplex.desc, prometheus.GaugeValue, float64(*value.TPhase)*0.1, host, eoj, channel, "t"))
						}
					}
				}
				if list := reading.CumulativeElectricEnergyListDuplex; list != nil {
					for i, value := range list.ElectricEnergy {
						channel := strconv.Itoa(int(list.StartChannel) + i)
						if value.NormalDirectionElectricEnergy != nil {
							metrics = append(metrics, prometheus.MustNewConstMetric(energyDuplex.desc, prometheus.CounterValue, float64(*value.NormalDirectionElectricEnergy)*scale, host, eoj, channel, "forward"))
						}
						if value.ReverseDirectionElectricEnergy != nil {
							metrics = append(metrics, prometheus.MustNewConstMetric(energyDuplex.desc, prometheus.CounterValue, float64(*value.ReverseDirectionElectricEnergy)*scale, host, eoj, channel, "reverse"))
						}
					}
				}
				if list := reading.InstantaneousCurrentListDuplex; list != nil {
					for i, value := range list.InstantaneousCurrent {
						channel := strconv.Itoa(int(list.StartChannel) + i)
						if value.RPhase != nil {
							metrics = append(metrics, prometheus.MustNewConstMetric(currentDuplex.desc, prometheus.GaugeValue, float64(*value.RPhase)*0.1, host, eoj, channel, "r"))
						}
						if value.TPhase != nil {
							metrics = append(metrics, prometheus.MustNewConstMetric(currentDuplex.desc, prometheus.GaugeValue, float64(*value.TPhase)*0.1, host, eoj, channel, "t"))
						}
					}
				}
				if list := reading.InstantaneousElectricPowerListDuplex; list != nil {
					for i, value := range list.InstantaneousElectricPower {
						if value != nil {
							channel := strconv.Itoa(int(list.StartChannel) + i)
							metrics = append(metrics, prometheus.MustNewConstMetric(powerDuplex.desc, prometheus.GaugeValue, float64(*value), host, eoj, channel))
						}
					}
				}
				if value := reading.InstantaneousCurrent; value != nil {
					if value.RPhase != nil {
						metrics = append(metrics, prometheus.MustNewConstMetric(current.desc, prometheus.GaugeValue, float64(*value.RPhase)*0.1, host, eoj, "r"))
					}
					if value.TPhase != nil {
						metrics = append(metrics, prometheus.MustNewConstMetric(current.desc, prometheus.GaugeValue, float64(*value.TPhase)*0.1, host, eoj, "t"))
					}
				}
				if value := reading.InstantaneousVoltage; value != nil {
					if value.RS != nil {
						metrics = append(metrics, prometheus.MustNewConstMetric(voltage.desc, prometheus.GaugeValue, float64(*value.RS)*0.1, host, eoj, "r_s"))
					}
					if value.ST != nil {
						metrics = append(metrics, prometheus.MustNewConstMetric(voltage.desc, prometheus.GaugeValue, float64(*value.ST)*0.1, host, eoj, "s_t"))
					}
				}

				channels := [32]*echonetlite.MeasurementChannel{
					reading.MeasurementChannel1, reading.MeasurementChannel2, reading.MeasurementChannel3, reading.MeasurementChannel4,
					reading.MeasurementChannel5, reading.MeasurementChannel6, reading.MeasurementChannel7, reading.MeasurementChannel8,
					reading.MeasurementChannel9, reading.MeasurementChannel10, reading.MeasurementChannel11, reading.MeasurementChannel12,
					reading.MeasurementChannel13, reading.MeasurementChannel14, reading.MeasurementChannel15, reading.MeasurementChannel16,
					reading.MeasurementChannel17, reading.MeasurementChannel18, reading.MeasurementChannel19, reading.MeasurementChannel20,
					reading.MeasurementChannel21, reading.MeasurementChannel22, reading.MeasurementChannel23, reading.MeasurementChannel24,
					reading.MeasurementChannel25, reading.MeasurementChannel26, reading.MeasurementChannel27, reading.MeasurementChannel28,
					reading.MeasurementChannel29, reading.MeasurementChannel30, reading.MeasurementChannel31, reading.MeasurementChannel32,
				}
				for i, value := range channels {
					if value == nil {
						continue
					}
					if value.ElectricEnergy != nil {
						metrics = append(metrics, prometheus.MustNewConstMetric(channelEnergy[i].desc, prometheus.CounterValue, float64(*value.ElectricEnergy)*scale, host, eoj))
					}
					if value.CurrentRPhase != nil {
						metrics = append(metrics, prometheus.MustNewConstMetric(channelCurrent[i].desc, prometheus.GaugeValue, float64(*value.CurrentRPhase)*0.1, host, eoj, "r"))
					}
					if value.CurrentTPhase != nil {
						metrics = append(metrics, prometheus.MustNewConstMetric(channelCurrent[i].desc, prometheus.GaugeValue, float64(*value.CurrentTPhase)*0.1, host, eoj, "t"))
					}
				}
				return superClass.appendSamples(metrics, reading.SuperClass, host, eoj)
			},
		),
	}
}

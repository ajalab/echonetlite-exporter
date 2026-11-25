package collector

import (
	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type superClassMetrics struct {
	operationStatus                       metricDescriptor
	instantaneousElectricPowerConsumption metricDescriptor
	consumedCumulativeElectricEnergy      metricDescriptor
	currentLimit                          metricDescriptor
	faultStatus                           metricDescriptor
	faultDescription                      metricDescriptor
	powerSaving                           metricDescriptor
	powerLimit                            metricDescriptor
	hourMeter                             metricDescriptor
}

func newSuperClassMetrics(prefix string) superClassMetrics {
	return superClassMetrics{
		operationStatus: newMetricDescriptor(echonetlite.EPCOperationStatus, "operationStatus", "", gaugeReading,
			prefix+"operation_status", "Operation status", deviceLabels),
		instantaneousElectricPowerConsumption: newMetricDescriptor(echonetlite.EPCInstantaneousElectricPowerConsumption, "instantaneousElectricPowerConsumption", "watts", instantaneousReading,
			prefix+"electric_power_consumption_watts", "Measured instantaneous power consumption", deviceLabels),
		consumedCumulativeElectricEnergy: newMetricDescriptor(echonetlite.EPCConsumedCumulativeElectricEnergy, "consumedCumulativeElectricEnergy", "joules", cumulativeReading,
			prefix+"consumed_electric_energy_joules_total", "Measured cumulative electric energy consumption", deviceLabels),
		currentLimit: newMetricDescriptor(echonetlite.EPCCurrentLimit, "currentLimit", "ratio", gaugeReading,
			prefix+"current_limit_ratio", "Current limit setting", deviceLabels),
		faultStatus: newMetricDescriptor(echonetlite.EPCFaultStatus, "faultStatus", "", gaugeReading,
			prefix+"fault_status", "Fault status", deviceLabels),
		faultDescription: newMetricDescriptor(echonetlite.EPCFaultDescription, "faultDescription", "", gaugeReading,
			prefix+"fault_description", "Fault description", deviceLabels),
		powerSaving: newMetricDescriptor(echonetlite.EPCPowerSaving, "powerSaving", "", gaugeReading,
			prefix+"power_saving", "Power-saving operation setting", deviceLabels),
		powerLimit: newMetricDescriptor(echonetlite.EPCPowerLimit, "powerLimit", "watts", gaugeReading,
			prefix+"power_limit_watts", "Power limit setting", deviceLabels),
		hourMeter: newMetricDescriptor(echonetlite.EPCHourMeter, "hourMeter", "seconds", cumulativeReading,
			prefix+"hour_meter_seconds_total", "Cumulative operating time", deviceLabels),
	}
}

func (m superClassMetrics) descriptors() []metricDescriptor {
	return []metricDescriptor{
		m.operationStatus, m.instantaneousElectricPowerConsumption, m.consumedCumulativeElectricEnergy,
		m.currentLimit, m.faultStatus, m.faultDescription, m.powerSaving, m.powerLimit, m.hourMeter,
	}
}

func (m superClassMetrics) appendSamples(metrics []prometheus.Metric, reading echonetlite.SuperClass, host, eoj string) []prometheus.Metric {
	if value := reading.OperationStatus; value != nil {
		metrics = append(metrics, prometheus.MustNewConstMetric(m.operationStatus.desc, prometheus.GaugeValue, float64(*value), host, eoj))
	}
	if value := reading.InstantaneousElectricPowerConsumption; value != nil {
		metrics = append(metrics, prometheus.MustNewConstMetric(m.instantaneousElectricPowerConsumption.desc, prometheus.GaugeValue, float64(*value), host, eoj))
	}
	if value := reading.ConsumedCumulativeElectricEnergy; value != nil {
		metrics = append(metrics, prometheus.MustNewConstMetric(m.consumedCumulativeElectricEnergy.desc, prometheus.CounterValue, float64(*value)*joulesPerWattHour, host, eoj))
	}
	if value := reading.CurrentLimit; value != nil {
		metrics = append(metrics, prometheus.MustNewConstMetric(m.currentLimit.desc, prometheus.GaugeValue, float64(*value)*0.01, host, eoj))
	}
	if value := reading.FaultStatus; value != nil {
		metrics = append(metrics, prometheus.MustNewConstMetric(m.faultStatus.desc, prometheus.GaugeValue, float64(*value), host, eoj))
	}
	if value := reading.FaultDescription; value != nil {
		metrics = append(metrics, prometheus.MustNewConstMetric(m.faultDescription.desc, prometheus.GaugeValue, float64(*value), host, eoj))
	}
	if value := reading.PowerSaving; value != nil {
		metrics = append(metrics, prometheus.MustNewConstMetric(m.powerSaving.desc, prometheus.GaugeValue, float64(*value), host, eoj))
	}
	if value := reading.PowerLimit; value != nil {
		metrics = append(metrics, prometheus.MustNewConstMetric(m.powerLimit.desc, prometheus.GaugeValue, float64(*value), host, eoj))
	}
	if value := reading.HourMeter; value != nil {
		metrics = append(metrics, prometheus.MustNewConstMetric(m.hourMeter.desc, prometheus.CounterValue, float64(*value), host, eoj))
	}
	return metrics
}

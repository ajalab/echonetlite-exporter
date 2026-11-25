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

const superClassOperationStatusValues = `{"48":"ON","49":"OFF"}`

func newSuperClassMetrics(prefix, operationStatusValues string) superClassMetrics {
	return superClassMetrics{
		operationStatus: newPropertyMetricDescriptorWithValues(echonetlite.EPCOperationStatus, "operationStatus",
			prefix+"operation_status", "Operation status", operationStatusValues),
		instantaneousElectricPowerConsumption: newPropertyMetricDescriptor(echonetlite.EPCInstantaneousElectricPowerConsumption, "instantaneousElectricPowerConsumption", "watts", instantaneousReading,
			prefix+"electric_power_consumption_watts", "Measured instantaneous power consumption", deviceLabels),
		consumedCumulativeElectricEnergy: newPropertyMetricDescriptor(echonetlite.EPCConsumedCumulativeElectricEnergy, "consumedCumulativeElectricEnergy", "joules", cumulativeReading,
			prefix+"consumed_electric_energy_joules_total", "Measured cumulative electric energy consumption", deviceLabels),
		currentLimit: newPropertyMetricDescriptor(echonetlite.EPCCurrentLimit, "currentLimit", "ratio", gaugeReading,
			prefix+"current_limit_ratio", "Current limit setting", deviceLabels),
		faultStatus: newPropertyMetricDescriptorWithValues(echonetlite.EPCFaultStatus, "faultStatus",
			prefix+"fault_status", "Fault status", `{"65":"Fault occurred.","66":"No fault has occurred."}`),
		faultDescription: newPropertyMetricDescriptorWithValues(echonetlite.EPCFaultDescription, "faultDescription",
			prefix+"fault_description", "Fault description", `{"0":"No fault","1":"Faults that can be recovered from by turn off the power or unplug and reoperate.","2":"Faults that can be recovered from by pressing the reset button.",`+
				`"3":"Device set incorrectly","4":"Supply","5":"Cleaning (filters, etc.)",`+
				`"6":"Changing the battery","7":"Recover operation no required","9":"User-definable domain 1",`+
				`"10...19":"Abnormal event or the tripping of a safety device.","20...29":"Fault in a switch","30...59":"Fault in the sensor system",`+
				`"60...89":"Fault in a component such as an actuator.","90...110":"Fault in a control circuit board.","111...1000":"User-definable domain 2",`+
				`"1001":"Repair location unknown","1023":"Fault"}`),
		powerSaving: newPropertyMetricDescriptorWithValues(echonetlite.EPCPowerSaving, "powerSaving",
			prefix+"power_saving", "Power-saving operation setting", `{"65":"Power saving Operation","66":"Normal Operation"}`),
		powerLimit: newPropertyMetricDescriptor(echonetlite.EPCPowerLimit, "powerLimit", "watts", gaugeReading,
			prefix+"power_limit_watts", "Power limit setting", deviceLabels),
		hourMeter: newPropertyMetricDescriptor(echonetlite.EPCHourMeter, "hourMeter", "seconds", cumulativeReading,
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

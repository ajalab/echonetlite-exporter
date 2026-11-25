package collector

import (
	"context"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// StorageBatteryCollector collects and publishes metrics for its ECHONET Lite device class.
type StorageBatteryCollector struct {
	*collector
}

// NewStorageBatteryCollector creates a StorageBatteryCollector.
func NewStorageBatteryCollector(
	get func(context.Context, string, echonetlite.EOJ) (*echonetlite.StorageBattery, error),
	interval time.Duration,
	targets []echonetlite.Device,
) *StorageBatteryCollector {
	chargeableEnergy := newMetricDescriptor(
		echonetlite.EPCACChargeableElectricEnergy,
		"acChargeableElectricEnergy", "joules", gaugeReading,
		"echonetlite_storage_battery_ac_chargeable_electric_energy_joules",
		"AC chargeable electric energy",
		deviceLabels,
	)
	dischargeableEnergy := newMetricDescriptor(
		echonetlite.EPCACDischargeableElectricEnergy,
		"acDischargeableElectricEnergy", "joules", gaugeReading,
		"echonetlite_storage_battery_ac_dischargeable_electric_energy_joules",
		"AC dischargeable electric energy",
		deviceLabels,
	)
	chargingEnergy := newMetricDescriptor(
		echonetlite.EPCACCumulativeChargingElectricEnergy,
		"acCumulativeChargingElectricEnergy", "joules", cumulativeReading,
		"echonetlite_storage_battery_ac_charging_electric_energy_joules_total",
		"AC measured cumulative charging electric energy",
		deviceLabels,
	)
	dischargingEnergy := newMetricDescriptor(
		echonetlite.EPCACCumulativeDischargingElectricEnergy,
		"acCumulativeDischargingElectricEnergy", "joules", cumulativeReading,
		"echonetlite_storage_battery_ac_discharging_electric_energy_joules_total",
		"AC measured cumulative discharging electric energy",
		deviceLabels,
	)
	boundsLabels := []string{"host", "eoj", "bound"}
	const prefix = "echonetlite_storage_battery_"
	acEffectiveChargingCapacity := newMetricDescriptor(0xA0, "acEffectiveChargingCapacity", "joules", gaugeReading,
		prefix+"ac_effective_charging_capacity_joules", "AC effective capacity (charging)", deviceLabels)
	acEffectiveDischargingCapacity := newMetricDescriptor(0xA1, "acEffectiveDischargingCapacity", "joules", gaugeReading,
		prefix+"ac_effective_discharging_capacity_joules", "AC effective capacity (discharging)", deviceLabels)
	acChargeableCapacity := newMetricDescriptor(0xA2, "acChargeableCapacity", "joules", gaugeReading,
		prefix+"ac_chargeable_capacity_joules", "AC chargeable capacity", deviceLabels)
	acDischargeableCapacity := newMetricDescriptor(0xA3, "acDischargeableCapacity", "joules", gaugeReading,
		prefix+"ac_dischargeable_capacity_joules", "AC dischargeable capacity", deviceLabels)
	acChargeUpperLimit := newMetricDescriptor(0xA6, "acChargeUpperLimit", "ratio", gaugeReading,
		prefix+"ac_charge_upper_limit_ratio", "AC charge upper limit setting", deviceLabels)
	acDischargeLowerLimit := newMetricDescriptor(0xA7, "acDischargeLowerLimit", "ratio", gaugeReading,
		prefix+"ac_discharge_lower_limit_ratio", "AC discharge lower limit setting", deviceLabels)
	acTargetChargingElectricEnergy := newMetricDescriptor(0xAA, "acTargetChargingElectricEnergy", "joules", gaugeReading,
		prefix+"ac_target_charging_electric_energy_joules", "AC charge amount setting value", deviceLabels)
	acTargetDischargingElectricEnergy := newMetricDescriptor(0xAB, "acTargetDischargingElectricEnergy", "joules", gaugeReading,
		prefix+"ac_target_discharging_electric_energy_joules", "AC discharge amount setting value", deviceLabels)
	chargingMethod := newMetricDescriptor(0xC1, "chargingMethod", "", gaugeReading,
		prefix+"charging_method", "Charging method", deviceLabels)
	dischargingMethod := newMetricDescriptor(0xC2, "dischargingMethod", "", gaugeReading,
		prefix+"discharging_method", "Discharging method", deviceLabels)
	acRatedElectricEnergy := newMetricDescriptor(0xC7, "acRatedElectricEnergy", "joules", gaugeReading,
		prefix+"ac_rated_electric_energy_joules", "AC rated electric energy", deviceLabels)
	minimumAndMaximumChargingElectricPower := newMetricDescriptor(0xC8, "minimumAndMaximumChargingElectricPower", "watts", gaugeReading,
		prefix+"minimum_and_maximum_charging_electric_power_watts", "Minimum/maximum charging electric power", boundsLabels)
	minimumAndMaximumDischargingElectricPower := newMetricDescriptor(0xC9, "minimumAndMaximumDischargingElectricPower", "watts", gaugeReading,
		prefix+"minimum_and_maximum_discharging_electric_power_watts", "Minimum/maximum discharging electric power", boundsLabels)
	minimumAndMaximumChargingCurrent := newMetricDescriptor(0xCA, "minimumAndMaximumChargingCurrent", "amperes", gaugeReading,
		prefix+"minimum_and_maximum_charging_current_amperes", "Minimum/maximum charging current", boundsLabels)
	minimumAndMaximumDischargingCurrent := newMetricDescriptor(0xCB, "minimumAndMaximumDischargingCurrent", "amperes", gaugeReading,
		prefix+"minimum_and_maximum_discharging_current_amperes", "Minimum/maximum discharging current", boundsLabels)
	reInterconnectionPermission := newMetricDescriptor(0xCC, "reInterconnectionPermission", "", gaugeReading,
		prefix+"re_interconnection_permission", "Re-interconnection permission setting", deviceLabels)
	operationPermission := newMetricDescriptor(0xCD, "operationPermission", "", gaugeReading,
		prefix+"operation_permission", "Operation permission setting", deviceLabels)
	independentOperationPermission := newMetricDescriptor(0xCE, "independentOperationPermission", "", gaugeReading,
		prefix+"independent_operation_permission", "Independent operation permission setting", deviceLabels)
	actualOperationMode := newMetricDescriptor(0xCF, "actualOperationMode", "", gaugeReading,
		prefix+"actual_operation_mode", "Working operation status", deviceLabels)
	ratedElectricEnergy := newMetricDescriptor(0xD0, "ratedElectricEnergy", "joules", gaugeReading,
		prefix+"rated_electric_energy_joules", "Rated electric energy", deviceLabels)
	ratedCapacity := newMetricDescriptor(0xD1, "ratedCapacity", "coulombs", gaugeReading,
		prefix+"rated_capacity_coulombs", "Rated capacity", deviceLabels)
	ratedVoltage := newMetricDescriptor(0xD2, "ratedVoltage", "volts", gaugeReading,
		prefix+"rated_voltage_volts", "Rated voltage", deviceLabels)
	instantaneousChargingAndDischargingElectricPower := newMetricDescriptor(0xD3, "instantaneousChargingAndDischargingElectricPower", "watts", instantaneousReading,
		prefix+"charging_and_discharging_electric_power_watts", "Measured instantaneous charging/discharging electric energy", deviceLabels)
	instantaneousChargingAndDischargingCurrent := newMetricDescriptor(0xD4, "instantaneousChargingAndDischargingCurrent", "amperes", instantaneousReading,
		prefix+"charging_and_discharging_current_amperes", "Measured instantaneous charging/discharging current", deviceLabels)
	instantaneousChargingAndDischargingVoltage := newMetricDescriptor(0xD5, "instantaneousChargingAndDischargingVoltage", "volts", instantaneousReading,
		prefix+"charging_and_discharging_voltage_volts", "Measured instantaneous charging/discharging voltage", deviceLabels)
	cumulativeDischargingElectricEnergy := newMetricDescriptor(0xD6, "cumulativeDischargingElectricEnergy", "joules", cumulativeReading,
		prefix+"discharging_electric_energy_joules_total", "Measured cumulative discharging electric energy", deviceLabels)
	cumulativeChargingElectricEnergy := newMetricDescriptor(0xD8, "cumulativeChargingElectricEnergy", "joules", cumulativeReading,
		prefix+"charging_electric_energy_joules_total", "Measured cumulative charging electric energy", deviceLabels)
	operationMode := newMetricDescriptor(0xDA, "operationMode", "", gaugeReading,
		prefix+"operation_mode", "Operation mode setting", deviceLabels)
	powerSystemInterconnectionStatus := newMetricDescriptor(0xDB, "powerSystemInterconnectionStatus", "", gaugeReading,
		prefix+"power_system_interconnection_status", "System-interconnected type", deviceLabels)
	minimumAndMaximumChargingPowerAtIndependent := newMetricDescriptor(0xDC, "minimumAndMaximumChargingPowerAtIndependent", "watts", gaugeReading,
		prefix+"minimum_and_maximum_charging_power_at_independent_watts", "Minimum/maximum charging power (Independent)", boundsLabels)
	minimumAndMaximumDischargingPowerAtIndependent := newMetricDescriptor(0xDD, "minimumAndMaximumDischargingPowerAtIndependent", "watts", gaugeReading,
		prefix+"minimum_and_maximum_discharging_power_at_independent_watts", "Minimum/maximum discharging power (Independent)", boundsLabels)
	minimumAndMaximumChargingCurrentAtIndependent := newMetricDescriptor(0xDE, "minimumAndMaximumChargingCurrentAtIndependent", "amperes", gaugeReading,
		prefix+"minimum_and_maximum_charging_current_at_independent_amperes", "Minimum/maximum charging current (Independent)", boundsLabels)
	minimumAndMaximumDischargingCurrentAtIndependent := newMetricDescriptor(0xDF, "minimumAndMaximumDischargingCurrentAtIndependent", "amperes", gaugeReading,
		prefix+"minimum_and_maximum_discharging_current_at_independent_amperes", "Minimum/maximum discharging current (Independent)", boundsLabels)
	remainingCapacity1 := newMetricDescriptor(0xE2, "remainingCapacity1", "joules", gaugeReading,
		prefix+"remaining_capacity1_joules", "Remaining stored electricity 1", deviceLabels)
	remainingCapacity2 := newMetricDescriptor(0xE3, "remainingCapacity2", "coulombs", gaugeReading,
		prefix+"remaining_capacity2_coulombs", "Remaining stored electricity 2", deviceLabels)
	remainingCapacity3 := newMetricDescriptor(0xE4, "remainingCapacity3", "ratio", gaugeReading,
		prefix+"remaining_capacity3_ratio", "Remaining stored electricity 3", deviceLabels)
	batteryHealthState := newMetricDescriptor(0xE5, "batteryHealthState", "ratio", gaugeReading,
		prefix+"battery_health_state_ratio", "Battery state of health", deviceLabels)
	superClass := newSuperClassMetrics(prefix)
	descs := []metricDescriptor{
		chargeableEnergy, dischargeableEnergy, chargingEnergy, dischargingEnergy,
		acEffectiveChargingCapacity,
		acEffectiveDischargingCapacity, acChargeableCapacity, acDischargeableCapacity, acChargeUpperLimit, acDischargeLowerLimit,
		acTargetChargingElectricEnergy, acTargetDischargingElectricEnergy, chargingMethod, dischargingMethod, acRatedElectricEnergy,
		minimumAndMaximumChargingElectricPower, minimumAndMaximumDischargingElectricPower, minimumAndMaximumChargingCurrent, minimumAndMaximumDischargingCurrent, reInterconnectionPermission,
		operationPermission, independentOperationPermission, actualOperationMode, ratedElectricEnergy, ratedCapacity,
		ratedVoltage, instantaneousChargingAndDischargingElectricPower, instantaneousChargingAndDischargingCurrent, instantaneousChargingAndDischargingVoltage, cumulativeDischargingElectricEnergy,
		cumulativeChargingElectricEnergy, operationMode, powerSystemInterconnectionStatus, minimumAndMaximumChargingPowerAtIndependent, minimumAndMaximumDischargingPowerAtIndependent,
		minimumAndMaximumChargingCurrentAtIndependent, minimumAndMaximumDischargingCurrentAtIndependent, remainingCapacity1, remainingCapacity2, remainingCapacity3,
		batteryHealthState,
	}
	descs = append(descs, superClass.descriptors()...)
	return &StorageBatteryCollector{
		collector: newCollector(interval, targets, descs, get,
			func(device echonetlite.Device, reading *echonetlite.StorageBattery) []prometheus.Metric {
				host, eoj := device.Host(), device.EOJ().String()
				metrics := []prometheus.Metric{
					prometheus.MustNewConstMetric(chargeableEnergy.desc, prometheus.GaugeValue, float64(reading.ACChargeableElectricEnergy)*joulesPerWattHour, host, eoj),
					prometheus.MustNewConstMetric(dischargeableEnergy.desc, prometheus.GaugeValue, float64(reading.ACDischargeableElectricEnergy)*joulesPerWattHour, host, eoj),
					prometheus.MustNewConstMetric(chargingEnergy.desc, prometheus.CounterValue, float64(reading.ACCumulativeChargingElectricEnergy)*joulesPerWattHour, host, eoj),
					prometheus.MustNewConstMetric(dischargingEnergy.desc, prometheus.CounterValue, float64(reading.ACCumulativeDischargingElectricEnergy)*joulesPerWattHour, host, eoj),
				}

				if value := reading.ACEffectiveChargingCapacity; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acEffectiveChargingCapacity.desc, prometheus.GaugeValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.ACEffectiveDischargingCapacity; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acEffectiveDischargingCapacity.desc, prometheus.GaugeValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.ACChargeableCapacity; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acChargeableCapacity.desc, prometheus.GaugeValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.ACDischargeableCapacity; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acDischargeableCapacity.desc, prometheus.GaugeValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.ACChargeUpperLimit; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acChargeUpperLimit.desc, prometheus.GaugeValue, float64(*value)*0.01, host, eoj))
				}
				if value := reading.ACDischargeLowerLimit; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acDischargeLowerLimit.desc, prometheus.GaugeValue, float64(*value)*0.01, host, eoj))
				}
				if value := reading.ACTargetChargingElectricEnergy; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acTargetChargingElectricEnergy.desc, prometheus.GaugeValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.ACTargetDischargingElectricEnergy; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acTargetDischargingElectricEnergy.desc, prometheus.GaugeValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.ChargingMethod; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(chargingMethod.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.DischargingMethod; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(dischargingMethod.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.ACRatedElectricEnergy; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(acRatedElectricEnergy.desc, prometheus.GaugeValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				bounds := [2]string{"min", "max"}
				for _, field := range []struct {
					desc   *prometheus.Desc
					values [2]*uint32
				}{
					{minimumAndMaximumChargingElectricPower.desc, reading.MinimumAndMaximumChargingElectricPower},
					{minimumAndMaximumDischargingElectricPower.desc, reading.MinimumAndMaximumDischargingElectricPower},
					{minimumAndMaximumChargingPowerAtIndependent.desc, reading.MinimumAndMaximumChargingPowerAtIndependent},
					{minimumAndMaximumDischargingPowerAtIndependent.desc, reading.MinimumAndMaximumDischargingPowerAtIndependent},
				} {
					for i, value := range field.values {
						if value != nil {
							metrics = append(metrics, prometheus.MustNewConstMetric(field.desc, prometheus.GaugeValue, float64(*value), host, eoj, bounds[i]))
						}
					}
				}
				for _, field := range []struct {
					desc   *prometheus.Desc
					values [2]*uint16
				}{
					{minimumAndMaximumChargingCurrent.desc, reading.MinimumAndMaximumChargingCurrent},
					{minimumAndMaximumDischargingCurrent.desc, reading.MinimumAndMaximumDischargingCurrent},
					{minimumAndMaximumChargingCurrentAtIndependent.desc, reading.MinimumAndMaximumChargingCurrentAtIndependent},
					{minimumAndMaximumDischargingCurrentAtIndependent.desc, reading.MinimumAndMaximumDischargingCurrentAtIndependent},
				} {
					for i, value := range field.values {
						if value != nil {
							metrics = append(metrics, prometheus.MustNewConstMetric(field.desc, prometheus.GaugeValue, float64(*value)*0.1, host, eoj, bounds[i]))
						}
					}
				}
				if value := reading.ReInterconnectionPermission; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(reInterconnectionPermission.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.OperationPermission; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(operationPermission.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.IndependentOperationPermission; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(independentOperationPermission.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.ActualOperationMode; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(actualOperationMode.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.RatedElectricEnergy; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(ratedElectricEnergy.desc, prometheus.GaugeValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.RatedCapacity; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(ratedCapacity.desc, prometheus.GaugeValue, float64(*value)*360, host, eoj))
				}
				if value := reading.RatedVoltage; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(ratedVoltage.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.InstantaneousChargingAndDischargingElectricPower; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(instantaneousChargingAndDischargingElectricPower.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.InstantaneousChargingAndDischargingCurrent; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(instantaneousChargingAndDischargingCurrent.desc, prometheus.GaugeValue, float64(*value)*0.1, host, eoj))
				}
				if value := reading.InstantaneousChargingAndDischargingVoltage; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(instantaneousChargingAndDischargingVoltage.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.CumulativeDischargingElectricEnergy; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(cumulativeDischargingElectricEnergy.desc, prometheus.CounterValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.CumulativeChargingElectricEnergy; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(cumulativeChargingElectricEnergy.desc, prometheus.CounterValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.OperationMode; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(operationMode.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.PowerSystemInterconnectionStatus; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(powerSystemInterconnectionStatus.desc, prometheus.GaugeValue, float64(*value), host, eoj))
				}
				if value := reading.RemainingCapacity1; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(remainingCapacity1.desc, prometheus.GaugeValue, float64(*value)*joulesPerWattHour, host, eoj))
				}
				if value := reading.RemainingCapacity2; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(remainingCapacity2.desc, prometheus.GaugeValue, float64(*value)*360, host, eoj))
				}
				if value := reading.RemainingCapacity3; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(remainingCapacity3.desc, prometheus.GaugeValue, float64(*value)*0.01, host, eoj))
				}
				if value := reading.BatteryHealthState; value != nil {
					metrics = append(metrics, prometheus.MustNewConstMetric(batteryHealthState.desc, prometheus.GaugeValue, float64(*value)*0.01, host, eoj))
				}
				return superClass.appendSamples(metrics, reading.SuperClass, host, eoj)
			},
		),
	}
}

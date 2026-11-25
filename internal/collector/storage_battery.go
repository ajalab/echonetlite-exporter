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
	chargeableEnergy := newPropertyMetricDescriptor(
		echonetlite.EPCACChargeableElectricEnergy,
		"acChargeableElectricEnergy", "joules", gaugeReading,
		"echonetlite_storage_battery_ac_chargeable_electric_energy_joules",
		"AC chargeable electric energy",
		deviceLabels,
	)
	dischargeableEnergy := newPropertyMetricDescriptor(
		echonetlite.EPCACDischargeableElectricEnergy,
		"acDischargeableElectricEnergy", "joules", gaugeReading,
		"echonetlite_storage_battery_ac_dischargeable_electric_energy_joules",
		"AC dischargeable electric energy",
		deviceLabels,
	)
	chargingEnergy := newPropertyMetricDescriptor(
		echonetlite.EPCACCumulativeChargingElectricEnergy,
		"acCumulativeChargingElectricEnergy", "joules", cumulativeReading,
		"echonetlite_storage_battery_ac_charging_electric_energy_joules_total",
		"AC measured cumulative charging electric energy",
		deviceLabels,
	)
	dischargingEnergy := newPropertyMetricDescriptor(
		echonetlite.EPCACCumulativeDischargingElectricEnergy,
		"acCumulativeDischargingElectricEnergy", "joules", cumulativeReading,
		"echonetlite_storage_battery_ac_discharging_electric_energy_joules_total",
		"AC measured cumulative discharging electric energy",
		deviceLabels,
	)
	boundsLabels := []string{"host", "eoj", "bound"}
	const prefix = "echonetlite_storage_battery_"
	acEffectiveChargingCapacity := newPropertyMetricDescriptor(0xA0, "acEffectiveChargingCapacity", "joules", gaugeReading,
		prefix+"ac_effective_charging_capacity_joules", "AC effective capacity (charging)", deviceLabels)
	acEffectiveDischargingCapacity := newPropertyMetricDescriptor(0xA1, "acEffectiveDischargingCapacity", "joules", gaugeReading,
		prefix+"ac_effective_discharging_capacity_joules", "AC effective capacity (discharging)", deviceLabels)
	acChargeableCapacity := newPropertyMetricDescriptor(0xA2, "acChargeableCapacity", "joules", gaugeReading,
		prefix+"ac_chargeable_capacity_joules", "AC chargeable capacity", deviceLabels)
	acDischargeableCapacity := newPropertyMetricDescriptor(0xA3, "acDischargeableCapacity", "joules", gaugeReading,
		prefix+"ac_dischargeable_capacity_joules", "AC dischargeable capacity", deviceLabels)
	acChargeUpperLimit := newPropertyMetricDescriptor(0xA6, "acChargeUpperLimit", "ratio", gaugeReading,
		prefix+"ac_charge_upper_limit_ratio", "AC charge upper limit setting", deviceLabels)
	acDischargeLowerLimit := newPropertyMetricDescriptor(0xA7, "acDischargeLowerLimit", "ratio", gaugeReading,
		prefix+"ac_discharge_lower_limit_ratio", "AC discharge lower limit setting", deviceLabels)
	acTargetChargingElectricEnergy := newPropertyMetricDescriptor(0xAA, "acTargetChargingElectricEnergy", "joules", gaugeReading,
		prefix+"ac_target_charging_electric_energy_joules", "AC charge amount setting value", deviceLabels)
	acTargetDischargingElectricEnergy := newPropertyMetricDescriptor(0xAB, "acTargetDischargingElectricEnergy", "joules", gaugeReading,
		prefix+"ac_target_discharging_electric_energy_joules", "AC discharge amount setting value", deviceLabels)
	chargingMethod := newPropertyMetricDescriptorWithValues(0xC1, "chargingMethod",
		prefix+"charging_method", "Charging method",
		`{"0":"Others","1":"Maximum charging electric energy charting","2":"Surplus electric energy charging",`+
			`"3":"Designated electric energy charging","4":"Designated current power charging"}`)
	dischargingMethod := newPropertyMetricDescriptorWithValues(0xC2, "dischargingMethod",
		prefix+"discharging_method", "Discharging method",
		`{"0":"Others","1":"Maximum discharge electric energy discharging","2":"Load following discharge",`+
			`"3":"Designated electric energy discharging","4":"Designated current power discharging"}`)
	acRatedElectricEnergy := newPropertyMetricDescriptor(0xC7, "acRatedElectricEnergy", "joules", gaugeReading,
		prefix+"ac_rated_electric_energy_joules", "AC rated electric energy", deviceLabels)
	minimumAndMaximumChargingElectricPower := newPropertyMetricDescriptor(0xC8, "minimumAndMaximumChargingElectricPower", "watts", gaugeReading,
		prefix+"minimum_and_maximum_charging_electric_power_watts", "Minimum/maximum charging electric power", boundsLabels)
	minimumAndMaximumDischargingElectricPower := newPropertyMetricDescriptor(0xC9, "minimumAndMaximumDischargingElectricPower", "watts", gaugeReading,
		prefix+"minimum_and_maximum_discharging_electric_power_watts", "Minimum/maximum discharging electric power", boundsLabels)
	minimumAndMaximumChargingCurrent := newPropertyMetricDescriptor(0xCA, "minimumAndMaximumChargingCurrent", "amperes", gaugeReading,
		prefix+"minimum_and_maximum_charging_current_amperes", "Minimum/maximum charging current", boundsLabels)
	minimumAndMaximumDischargingCurrent := newPropertyMetricDescriptor(0xCB, "minimumAndMaximumDischargingCurrent", "amperes", gaugeReading,
		prefix+"minimum_and_maximum_discharging_current_amperes", "Minimum/maximum discharging current", boundsLabels)
	const permissionValues = `{"65":"Permitted","66":"Prohibited"}`
	reInterconnectionPermission := newPropertyMetricDescriptorWithValues(0xCC, "reInterconnectionPermission",
		prefix+"re_interconnection_permission", "Re-interconnection permission setting", permissionValues)
	operationPermission := newPropertyMetricDescriptorWithValues(0xCD, "operationPermission",
		prefix+"operation_permission", "Operation permission setting", permissionValues)
	independentOperationPermission := newPropertyMetricDescriptorWithValues(0xCE, "independentOperationPermission",
		prefix+"independent_operation_permission", "Independent operation permission setting", permissionValues)
	const operationModeValues = `{"64":"Other","65":"Rapid charging","66":"Charging","67":"Discharging",` +
		`"68":"Standby","69":"Test","70":"Automatic","72":"Restart","73":"Effective capacity recalculation processing"}`
	actualOperationMode := newPropertyMetricDescriptorWithValues(0xCF, "actualOperationMode",
		prefix+"actual_operation_mode", "Working operation status", operationModeValues)
	ratedElectricEnergy := newPropertyMetricDescriptor(0xD0, "ratedElectricEnergy", "joules", gaugeReading,
		prefix+"rated_electric_energy_joules", "Rated electric energy", deviceLabels)
	ratedCapacity := newPropertyMetricDescriptor(0xD1, "ratedCapacity", "coulombs", gaugeReading,
		prefix+"rated_capacity_coulombs", "Rated capacity", deviceLabels)
	ratedVoltage := newPropertyMetricDescriptor(0xD2, "ratedVoltage", "volts", gaugeReading,
		prefix+"rated_voltage_volts", "Rated voltage", deviceLabels)
	instantaneousChargingAndDischargingElectricPower := newPropertyMetricDescriptor(0xD3, "instantaneousChargingAndDischargingElectricPower", "watts", instantaneousReading,
		prefix+"charging_and_discharging_electric_power_watts", "Measured instantaneous charging/discharging electric energy", deviceLabels)
	instantaneousChargingAndDischargingCurrent := newPropertyMetricDescriptor(0xD4, "instantaneousChargingAndDischargingCurrent", "amperes", instantaneousReading,
		prefix+"charging_and_discharging_current_amperes", "Measured instantaneous charging/discharging current", deviceLabels)
	instantaneousChargingAndDischargingVoltage := newPropertyMetricDescriptor(0xD5, "instantaneousChargingAndDischargingVoltage", "volts", instantaneousReading,
		prefix+"charging_and_discharging_voltage_volts", "Measured instantaneous charging/discharging voltage", deviceLabels)
	cumulativeDischargingElectricEnergy := newPropertyMetricDescriptor(0xD6, "cumulativeDischargingElectricEnergy", "joules", cumulativeReading,
		prefix+"discharging_electric_energy_joules_total", "Measured cumulative discharging electric energy", deviceLabels)
	cumulativeChargingElectricEnergy := newPropertyMetricDescriptor(0xD8, "cumulativeChargingElectricEnergy", "joules", cumulativeReading,
		prefix+"charging_electric_energy_joules_total", "Measured cumulative charging electric energy", deviceLabels)
	operationMode := newPropertyMetricDescriptorWithValues(0xDA, "operationMode",
		prefix+"operation_mode", "Operation mode setting", operationModeValues)
	powerSystemInterconnectionStatus := newPropertyMetricDescriptorWithValues(0xDB, "powerSystemInterconnectionStatus",
		prefix+"power_system_interconnection_status", "System-interconnected type", `{"0":"System interconnection (reverse power flow acceptable)","1":"Independent type","2":"System-interconnected type (reverse power flow not acceptable)"}`)
	minimumAndMaximumChargingPowerAtIndependent := newPropertyMetricDescriptor(0xDC, "minimumAndMaximumChargingPowerAtIndependent", "watts", gaugeReading,
		prefix+"minimum_and_maximum_charging_power_at_independent_watts", "Minimum/maximum charging power (Independent)", boundsLabels)
	minimumAndMaximumDischargingPowerAtIndependent := newPropertyMetricDescriptor(0xDD, "minimumAndMaximumDischargingPowerAtIndependent", "watts", gaugeReading,
		prefix+"minimum_and_maximum_discharging_power_at_independent_watts", "Minimum/maximum discharging power (Independent)", boundsLabels)
	minimumAndMaximumChargingCurrentAtIndependent := newPropertyMetricDescriptor(0xDE, "minimumAndMaximumChargingCurrentAtIndependent", "amperes", gaugeReading,
		prefix+"minimum_and_maximum_charging_current_at_independent_amperes", "Minimum/maximum charging current (Independent)", boundsLabels)
	minimumAndMaximumDischargingCurrentAtIndependent := newPropertyMetricDescriptor(0xDF, "minimumAndMaximumDischargingCurrentAtIndependent", "amperes", gaugeReading,
		prefix+"minimum_and_maximum_discharging_current_at_independent_amperes", "Minimum/maximum discharging current (Independent)", boundsLabels)
	remainingCapacity1 := newPropertyMetricDescriptor(0xE2, "remainingCapacity1", "joules", gaugeReading,
		prefix+"remaining_capacity1_joules", "Remaining stored electricity 1", deviceLabels)
	remainingCapacity2 := newPropertyMetricDescriptor(0xE3, "remainingCapacity2", "coulombs", gaugeReading,
		prefix+"remaining_capacity2_coulombs", "Remaining stored electricity 2", deviceLabels)
	remainingCapacity3 := newPropertyMetricDescriptor(0xE4, "remainingCapacity3", "ratio", gaugeReading,
		prefix+"remaining_capacity3_ratio", "Remaining stored electricity 3", deviceLabels)
	batteryHealthState := newPropertyMetricDescriptor(0xE5, "batteryHealthState", "ratio", gaugeReading,
		prefix+"battery_health_state_ratio", "Battery state of health", deviceLabels)
	superClass := newSuperClassMetrics(prefix, superClassOperationStatusValues)
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

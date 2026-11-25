package collector

import (
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
)

func TestPowerDistributionBoardMeteringCollectorPublishesMetrics(t *testing.T) {
	reading := &echonetlite.PowerDistributionBoardMetering{
		SuperClass:                      echonetlite.SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		UnitForCumulativeElectricEnergy: 0.1,
		InstantaneousCurrentListSimplex: &echonetlite.InstantaneousCurrentList{
			StartChannel: 3, InstantaneousCurrent: []echonetlite.CurrentPhases{{RPhase: new(int16(-15))}},
		},
		CumulativeElectricEnergyListDuplex: &echonetlite.CumulativeElectricEnergyListDuplex{
			StartChannel: 4, ElectricEnergy: []echonetlite.DirectionalEnergy{{ReverseDirectionElectricEnergy: new(uint32(2))}},
		},
		InstantaneousVoltage: &echonetlite.LineVoltage{ST: new(uint16(1000))},
		MeasurementChannel32: &echonetlite.MeasurementChannel{ElectricEnergy: new(uint32(1)), CurrentTPhase: new(int16(25))},
		InstantaneousElectricPowerListSimplex: echonetlite.InstantaneousElectricPowerListSimplex{
			StartChannel: 3, InstantaneousElectricPower: []*int32{new(int32(100)), nil, new(int32(-60))},
		},
		CumulativeElectricEnergyListSimplex: echonetlite.CumulativeElectricEnergyListSimplex{
			StartChannel: 3, ElectricEnergy: []*uint32{new(uint32(2))},
		},
	}
	const want = `
# HELP echonetlite_power_distribution_board_metering_electric_energy_list_simplex_joules_total Measured cumulative amount of electric power consumption list (simplex). EPC=0xB3
# TYPE echonetlite_power_distribution_board_metering_electric_energy_list_simplex_joules_total counter
echonetlite_power_distribution_board_metering_electric_energy_list_simplex_joules_total{channel="3",eoj="028701",host="192.0.2.10"} 720000
# HELP echonetlite_power_distribution_board_metering_electric_power_list_simplex_watts Measured instantaneous power consumption list (simplex). EPC=0xB7
# TYPE echonetlite_power_distribution_board_metering_electric_power_list_simplex_watts gauge
echonetlite_power_distribution_board_metering_electric_power_list_simplex_watts{channel="3",eoj="028701",host="192.0.2.10"} 100
echonetlite_power_distribution_board_metering_electric_power_list_simplex_watts{channel="5",eoj="028701",host="192.0.2.10"} -60
# HELP echonetlite_power_distribution_board_metering_current_list_simplex_amperes Measured instantaneous current list (simplex). EPC=0xB5
# TYPE echonetlite_power_distribution_board_metering_current_list_simplex_amperes gauge
echonetlite_power_distribution_board_metering_current_list_simplex_amperes{channel="3",eoj="028701",host="192.0.2.10",phase="r"} -1.5
# HELP echonetlite_power_distribution_board_metering_electric_energy_list_duplex_joules_total Measured cumulative amount of electric power consumption list (duplex). EPC=0xBA
# TYPE echonetlite_power_distribution_board_metering_electric_energy_list_duplex_joules_total counter
echonetlite_power_distribution_board_metering_electric_energy_list_duplex_joules_total{channel="4",direction="reverse",eoj="028701",host="192.0.2.10"} 720000
# HELP echonetlite_power_distribution_board_metering_measurement_channel32_amperes Measurement channel 32. EPC=0xEF
# TYPE echonetlite_power_distribution_board_metering_measurement_channel32_amperes gauge
echonetlite_power_distribution_board_metering_measurement_channel32_amperes{eoj="028701",host="192.0.2.10",phase="t"} 2.5
# HELP echonetlite_power_distribution_board_metering_measurement_channel32_joules_total Measurement channel 32. EPC=0xEF
# TYPE echonetlite_power_distribution_board_metering_measurement_channel32_joules_total counter
echonetlite_power_distribution_board_metering_measurement_channel32_joules_total{eoj="028701",host="192.0.2.10"} 360000
# HELP echonetlite_power_distribution_board_metering_voltage_volts Measured instantaneous voltages. EPC=0xC8
# TYPE echonetlite_power_distribution_board_metering_voltage_volts gauge
echonetlite_power_distribution_board_metering_voltage_volts{eoj="028701",host="192.0.2.10",line_pair="s_t"} 100
# HELP echonetlite_power_distribution_board_metering_fault_status Fault status. values={"65":"Fault occurred.","66":"No fault has occurred."} EPC=0x88
# TYPE echonetlite_power_distribution_board_metering_fault_status gauge
echonetlite_power_distribution_board_metering_fault_status{eoj="028701",host="192.0.2.10"} 66
# HELP echonetlite_power_distribution_board_metering_operation_status Operation status. values={"48":"ON","49":"OFF"} EPC=0x80
# TYPE echonetlite_power_distribution_board_metering_operation_status gauge
echonetlite_power_distribution_board_metering_operation_status{eoj="028701",host="192.0.2.10"} 48
`
	assertPublishedMetrics(t, NewPowerDistributionBoardMeteringCollector, echonetlite.ClassPowerDistributionBoardMetering, reading, want)
}

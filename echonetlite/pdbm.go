package echonetlite

import "context"

const (
	epcCumulativeElectricEnergyListSimplex   = 0xB3
	epcInstantaneousElectricPowerListSimplex = 0xB7
	epcUnitForCumulativeElectricEnergy       = 0xC2
)

type PowerDistributionBoardMetering struct {
	host string
	eoj  EOJ
	conn *Connection
}

func NewPowerDistributionBoardMetering(host string, eoj EOJ, conn *Connection) *PowerDistributionBoardMetering {
	return &PowerDistributionBoardMetering{
		host: host,
		eoj:  eoj,
		conn: conn,
	}
}

func (p *PowerDistributionBoardMetering) Host() string {
	return p.host
}

func (p *PowerDistributionBoardMetering) EOJ() EOJ {
	return p.eoj
}

func (p *PowerDistributionBoardMetering) Get(ctx context.Context) (*PowerDistributionBoardMeteringProps, error) {
	getReq := Frame{
		SEOJ: EOJ{0x05, 0xFF, 0x01},
		DEOJ: EOJ{0x02, 0x87, 0x01},
		ESV:  0x62,
		Properties: []Property{
			{EPC: epcCumulativeElectricEnergyListSimplex, EDT: []byte{}},
			{EPC: epcInstantaneousElectricPowerListSimplex, EDT: []byte{}},
			{EPC: epcUnitForCumulativeElectricEnergy, EDT: []byte{}},
		},
	}
	resFrame, err := p.conn.unicast(ctx, p.host, getReq)
	if err != nil {
		return nil, err
	}

	var instantaneousElectricPowerListSimplex InstantaneousElectricPowerListSimplex
	var cumulativeElectricEnergyListSimplex CumulativeElectricEnergyListSimplex
	var unitForCumulativeEnergy float32
	for _, p := range resFrame.Properties {
		switch p.EPC {
		case epcInstantaneousElectricPowerListSimplex:
			instantaneousElectricPowerListSimplex = parseInstantaneousPowerListSimplex(p.EDT)
		case epcCumulativeElectricEnergyListSimplex:
			cumulativeElectricEnergyListSimplex = parseCumulativeElectricEnergyListSimplex(p.EDT)
		case epcUnitForCumulativeElectricEnergy:
			unitForCumulativeEnergy = parseUnitForCumulativeEnergy(p.EDT)
		}
	}

	return &PowerDistributionBoardMeteringProps{
		InstantaneousElectricPowerListSimplex: instantaneousElectricPowerListSimplex,
		CumulativeElectricEnergyListSimplex:   cumulativeElectricEnergyListSimplex,
		UnitForCumulativeEnergy:               unitForCumulativeEnergy,
	}, nil
}

type PowerDistributionBoardMeteringProps struct {
	// This property indicates the measured cumulative amount of electric power
	// consumption of a measurement channel specified by the property of 'Channel
	// range specification for cumulative amount of electric power consumption
	// measurement (simplex).
	CumulativeElectricEnergyListSimplex CumulativeElectricEnergyListSimplex

	// This property indicates the measured instantaneous power consumption of a
	// measurement channel specified by the property of 'Channel range
	// specification for instantaneous power consumption measurement (simplex).
	InstantaneousElectricPowerListSimplex InstantaneousElectricPowerListSimplex

	// This property indicates the unit (multiplying factor) used for the
	// measured cumulative amount of electric energy and the historical data of
	// measured cumulative amounts of electric energy.
	UnitForCumulativeEnergy float32
}

type InstantaneousElectricPowerListSimplex struct {
	StartChannel               uint8
	Range                      uint8
	InstantaneousElectricPower []int32
}

func parseInstantaneousPowerListSimplex(edts []byte) InstantaneousElectricPowerListSimplex {
	if len(edts) < 2 {
		return InstantaneousElectricPowerListSimplex{}
	}
	startChannel := edts[0]
	channelRange := edts[1]
	valuesBytes := edts[2:]

	var instantaneousPowerList []int32
	for i := 0; i+4 <= len(valuesBytes); i += 4 {
		val := int32(valuesBytes[i])<<24 | int32(valuesBytes[i+1])<<16 | int32(valuesBytes[i+2])<<8 | int32(valuesBytes[i+3])
		instantaneousPowerList = append(instantaneousPowerList, val)
	}
	return InstantaneousElectricPowerListSimplex{
		StartChannel:               startChannel,
		Range:                      channelRange,
		InstantaneousElectricPower: instantaneousPowerList,
	}
}

type CumulativeElectricEnergyListSimplex struct {
	StartChannel   uint8
	Range          uint8
	ElectricEnergy []int32
}

func parseCumulativeElectricEnergyListSimplex(edts []byte) CumulativeElectricEnergyListSimplex {
	if len(edts) < 2 {
		return CumulativeElectricEnergyListSimplex{}
	}
	startChannel := edts[0]
	channelRange := edts[1]
	valuesBytes := edts[2:]

	var energyList []int32
	for i := 0; i+4 <= len(valuesBytes); i += 4 {
		val := int32(valuesBytes[i])<<24 | int32(valuesBytes[i+1])<<16 | int32(valuesBytes[i+2])<<8 | int32(valuesBytes[i+3])
		energyList = append(energyList, val)
	}
	return CumulativeElectricEnergyListSimplex{
		StartChannel:   startChannel,
		Range:          channelRange,
		ElectricEnergy: energyList,
	}
}

func parseUnitForCumulativeEnergy(edts []byte) float32 {
	{
		if len(edts) != 1 {
			return 1.0
		}
		switch edts[0] {
		case 0x00:
			return 1.0
		case 0x01:
			return 0.1
		case 0x02:
			return 0.01
		case 0x03:
			return 0.001
		case 0x04:
			return 0.0001
		case 0x0A:
			return 10.0
		case 0x0B:
			return 100.0
		case 0x0C:
			return 1000.0
		case 0x0D:
			return 10000.0
		default:
			return 1.0
		}
	}
}

package echonetlite

import (
	"context"
	"fmt"
)

const (
	EPCCumulativeElectricEnergyListSimplex   = 0xB3
	EPCInstantaneousElectricPowerListSimplex = 0xB7
	epcUnitForCumulativeElectricEnergy       = 0xC2
)

type PowerDistributionBoardMeteringClient struct {
	conn *Connection
}

func NewPowerDistributionBoardMeteringClient(conn *Connection) *PowerDistributionBoardMeteringClient {
	return &PowerDistributionBoardMeteringClient{conn: conn}
}

type PowerDistributionBoardMetering struct {
	CumulativeElectricEnergyListSimplex   CumulativeElectricEnergyListSimplex
	InstantaneousElectricPowerListSimplex InstantaneousElectricPowerListSimplex
	UnitForCumulativeElectricEnergy       float64 // kWh per count
}

type InstantaneousElectricPowerListSimplex struct {
	StartChannel               uint8
	InstantaneousElectricPower []*int32 // nil entries denote channels with NoData
}

type CumulativeElectricEnergyListSimplex struct {
	StartChannel   uint8
	ElectricEnergy []*uint32 // nil entries denote channels with NoData
}

func (p *PowerDistributionBoardMeteringClient) Get(ctx context.Context, host string, eoj EOJ) (*PowerDistributionBoardMetering, error) {
	read := propertyReader{conn: p.conn, host: host, eoj: eoj}
	supported, err := read.readableProperties(ctx)
	if err != nil {
		return nil, err
	}
	unitValues, err := read.required(ctx, epcUnitForCumulativeElectricEnergy)
	if err != nil {
		return nil, err
	}
	unit, err := decodeElectricEnergyUnit(unitValues[epcUnitForCumulativeElectricEnergy])
	if err != nil {
		return nil, fmt.Errorf("unitForCumulativeElectricEnergy (EPC 0xC2): %w", err)
	}
	values, err := read.optionalIndividually(ctx, supportedEPCs(supported,
		EPCCumulativeElectricEnergyListSimplex, EPCInstantaneousElectricPowerListSimplex))
	if err != nil {
		return nil, err
	}

	reading := &PowerDistributionBoardMetering{UnitForCumulativeElectricEnergy: unit}
	if edt := values[EPCCumulativeElectricEnergyListSimplex]; len(edt) != 0 {
		start, entries, err := decodeChannelList(edt, noEnergyData, decodeChannelEnergy)
		if err != nil {
			return nil, fmt.Errorf("cumulativeElectricEnergyListSimplex (EPC 0xB3): %w", err)
		}
		reading.CumulativeElectricEnergyListSimplex = CumulativeElectricEnergyListSimplex{StartChannel: start, ElectricEnergy: entries}
	}
	if edt := values[EPCInstantaneousElectricPowerListSimplex]; len(edt) != 0 {
		start, entries, err := decodeChannelList(edt, noPowerData, decodeElectricPower)
		if err != nil {
			return nil, fmt.Errorf("instantaneousElectricPowerListSimplex (EPC 0xB7): %w", err)
		}
		reading.InstantaneousElectricPowerListSimplex = InstantaneousElectricPowerListSimplex{StartChannel: start, InstantaneousElectricPower: entries}
	}
	return reading, nil
}

func decodeChannelEnergy(edt []byte) (uint32, error) {
	return decodeMeterValue(edt, 99_999_999)
}

func decodeChannelList[T int32 | uint32](edt []byte, noData uint32, decode func([]byte) (T, error)) (uint8, []*T, error) {
	if len(edt) < 2 {
		return 0, nil, fmt.Errorf("missing channel range")
	}
	start, count := edt[0], int(edt[1])
	// The Appendix specifies an unset range followed by one NoData entry at startup.
	if start == 0xFD && count == 0xFD && isNoData(edt[2:], noData) {
		return 0, nil, nil
	}
	if start == 0 || count < 1 || count > 60 || int(start)+count-1 > 252 {
		return 0, nil, fmt.Errorf("invalid channel range: start %d, count %d", start, count)
	}
	if len(edt) != 2+4*count {
		return 0, nil, fmt.Errorf("expected %d bytes for %d channels, got %d", 2+4*count, count, len(edt))
	}
	entries := make([]*T, count)
	for i := range entries {
		valueBytes := edt[2+4*i : 2+4*(i+1)]
		if isNoData(valueBytes, noData) {
			continue
		}
		value, err := decode(valueBytes)
		if err != nil {
			return 0, nil, fmt.Errorf("channel %d: %w", int(start)+i, err)
		}
		entries[i] = &value
	}
	return start, entries, nil
}

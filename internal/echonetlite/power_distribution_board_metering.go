package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
	"maps"
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
	SuperClass

	CumulativeElectricEnergyListSimplex   CumulativeElectricEnergyListSimplex
	InstantaneousElectricPowerListSimplex InstantaneousElectricPowerListSimplex
	UnitForCumulativeElectricEnergy       float64 // kWh per count

	MasterRatedCapacity                      *uint8 // A
	NumberOfMeasurementChannelsSimplex       *uint8
	NumberOfMeasurementChannelsDuplex        *uint8
	InstantaneousCurrentListSimplex          *InstantaneousCurrentList
	CumulativeElectricEnergyListDuplex       *CumulativeElectricEnergyListDuplex
	InstantaneousCurrentListDuplex           *InstantaneousCurrentList
	InstantaneousElectricPowerListDuplex     *InstantaneousElectricPowerListDuplex
	NormalDirectionCumulativeElectricEnergy  *uint32
	ReverseDirectionCumulativeElectricEnergy *uint32
	InstantaneousElectricPower               *int32
	InstantaneousCurrent                     *CurrentPhases
	InstantaneousVoltage                     *LineVoltage

	MeasurementChannel1  *MeasurementChannel
	MeasurementChannel2  *MeasurementChannel
	MeasurementChannel3  *MeasurementChannel
	MeasurementChannel4  *MeasurementChannel
	MeasurementChannel5  *MeasurementChannel
	MeasurementChannel6  *MeasurementChannel
	MeasurementChannel7  *MeasurementChannel
	MeasurementChannel8  *MeasurementChannel
	MeasurementChannel9  *MeasurementChannel
	MeasurementChannel10 *MeasurementChannel
	MeasurementChannel11 *MeasurementChannel
	MeasurementChannel12 *MeasurementChannel
	MeasurementChannel13 *MeasurementChannel
	MeasurementChannel14 *MeasurementChannel
	MeasurementChannel15 *MeasurementChannel
	MeasurementChannel16 *MeasurementChannel
	MeasurementChannel17 *MeasurementChannel
	MeasurementChannel18 *MeasurementChannel
	MeasurementChannel19 *MeasurementChannel
	MeasurementChannel20 *MeasurementChannel
	MeasurementChannel21 *MeasurementChannel
	MeasurementChannel22 *MeasurementChannel
	MeasurementChannel23 *MeasurementChannel
	MeasurementChannel24 *MeasurementChannel
	MeasurementChannel25 *MeasurementChannel
	MeasurementChannel26 *MeasurementChannel
	MeasurementChannel27 *MeasurementChannel
	MeasurementChannel28 *MeasurementChannel
	MeasurementChannel29 *MeasurementChannel
	MeasurementChannel30 *MeasurementChannel
	MeasurementChannel31 *MeasurementChannel
	MeasurementChannel32 *MeasurementChannel
}

type InstantaneousElectricPowerListSimplex struct {
	StartChannel               uint8
	InstantaneousElectricPower []*int32 // nil entries denote channels without a measurement
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

	monitoringRequired, monitoringOptional := superClassEPCs(supported, 0xC0, 0xC1)
	monitoringOptional = append(monitoringOptional, supportedEPCs(supported, 0xB0, 0xB1, 0xB8, 0xC6, 0xC7, 0xC8)...)
	for epc := 0xD0; epc <= 0xEF; epc++ {
		if supported[epc] {
			monitoringOptional = append(monitoringOptional, uint8(epc))
		}
	}
	monitoringValues, err := read.group(ctx, monitoringRequired, monitoringOptional)
	if err != nil {
		return nil, err
	}
	individual := supportedEPCs(supported, 0xB5, 0xBA, 0xBC, 0xBE)
	// Retry optional EPCs omitted from a short Get_SNA.
	for _, epc := range monitoringOptional {
		if _, present := monitoringValues[epc]; !present {
			individual = append(individual, epc)
		}
	}
	if len(individual) != 0 {
		additional, err := read.optionalIndividually(ctx, individual)
		if err != nil {
			return nil, err
		}
		maps.Copy(monitoringValues, additional)
	}

	reading.SuperClass, err = decodeSuperClass(monitoringValues)
	if err != nil {
		return nil, err
	}

	for _, field := range []struct {
		epc      uint8
		min, max uint8
		dest     **uint8
		noData   bool
	}{
		{0xB0, 0, 253, &reading.MasterRatedCapacity, false},
		{0xB1, 1, 252, &reading.NumberOfMeasurementChannelsSimplex, true},
		{0xB8, 1, 252, &reading.NumberOfMeasurementChannelsDuplex, true},
	} {
		edt := monitoringValues[field.epc]
		if len(edt) == 0 || (field.noData && len(edt) == 1 && edt[0] == 0xFD) {
			continue
		}
		value, err := decodeByte(edt, field.min, field.max)
		if err != nil {
			return nil, fmt.Errorf("EPC 0x%02X: %w", field.epc, err)
		}
		*field.dest = &value
	}
	for _, field := range []struct {
		epc  uint8
		dest **uint32
	}{
		{0xC0, &reading.NormalDirectionCumulativeElectricEnergy},
		{0xC1, &reading.ReverseDirectionCumulativeElectricEnergy},
	} {
		edt := monitoringValues[field.epc]
		if isNoData(edt, noEnergyData) {
			continue
		}
		value, err := decodeChannelEnergy(edt)
		if err != nil {
			return nil, fmt.Errorf("EPC 0x%02X: %w", field.epc, err)
		}
		*field.dest = &value
	}
	if edt := monitoringValues[0xB5]; len(edt) != 0 {
		value, err := decodeBoardCurrentList(edt)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xB5: %w", err)
		}
		reading.InstantaneousCurrentListSimplex = &value
	}
	if edt := monitoringValues[0xBA]; len(edt) != 0 {
		value, err := decodeBoardEnergyListDuplex(edt)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xBA: %w", err)
		}
		reading.CumulativeElectricEnergyListDuplex = &value
	}
	if edt := monitoringValues[0xBC]; len(edt) != 0 {
		value, err := decodeBoardCurrentList(edt)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xBC: %w", err)
		}
		reading.InstantaneousCurrentListDuplex = &value
	}
	if edt := monitoringValues[0xBE]; len(edt) != 0 {
		start, values, err := decodeChannelList(edt, noPowerData, decodeElectricPower)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xBE: %w", err)
		}
		reading.InstantaneousElectricPowerListDuplex = &InstantaneousElectricPowerListDuplex{
			StartChannel: start, InstantaneousElectricPower: values,
		}
	}
	if edt := monitoringValues[0xC6]; len(edt) != 0 && !isNoData(edt, noPowerData) && !isBoardPowerOutOfRange(edt) {
		value, err := decodeElectricPower(edt)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xC6: %w", err)
		}
		reading.InstantaneousElectricPower = &value
	}
	if edt := monitoringValues[0xC7]; len(edt) != 0 {
		value, err := decodeCurrentPhases(edt)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xC7: %w", err)
		}
		reading.InstantaneousCurrent = &value
	}
	if edt := monitoringValues[0xC8]; len(edt) != 0 {
		value, err := decodeLineVoltage(edt)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xC8: %w", err)
		}
		reading.InstantaneousVoltage = &value
	}
	legacyChannels := [32]**MeasurementChannel{
		&reading.MeasurementChannel1, &reading.MeasurementChannel2, &reading.MeasurementChannel3, &reading.MeasurementChannel4,
		&reading.MeasurementChannel5, &reading.MeasurementChannel6, &reading.MeasurementChannel7, &reading.MeasurementChannel8,
		&reading.MeasurementChannel9, &reading.MeasurementChannel10, &reading.MeasurementChannel11, &reading.MeasurementChannel12,
		&reading.MeasurementChannel13, &reading.MeasurementChannel14, &reading.MeasurementChannel15, &reading.MeasurementChannel16,
		&reading.MeasurementChannel17, &reading.MeasurementChannel18, &reading.MeasurementChannel19, &reading.MeasurementChannel20,
		&reading.MeasurementChannel21, &reading.MeasurementChannel22, &reading.MeasurementChannel23, &reading.MeasurementChannel24,
		&reading.MeasurementChannel25, &reading.MeasurementChannel26, &reading.MeasurementChannel27, &reading.MeasurementChannel28,
		&reading.MeasurementChannel29, &reading.MeasurementChannel30, &reading.MeasurementChannel31, &reading.MeasurementChannel32,
	}
	for i, channel := range legacyChannels {
		epc := 0xD0 + uint8(i)
		edt := monitoringValues[epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeMeasurementChannel(edt)
		if err != nil {
			return nil, fmt.Errorf("EPC 0x%02X: %w", epc, err)
		}
		*channel = &value
	}
	return reading, nil
}

func decodeChannelEnergy(edt []byte) (uint32, error) {
	return decodeMeterValue(edt, 99_999_999)
}

func isBoardPowerOutOfRange(edt []byte) bool {
	return isNoData(edt, 0x80000000) || isNoData(edt, 0x7FFFFFFF)
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
		if isNoData(valueBytes, noData) || (noData == noPowerData && isBoardPowerOutOfRange(valueBytes)) {
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

type InstantaneousCurrentList struct {
	StartChannel         uint8
	InstantaneousCurrent []CurrentPhases
}

type CumulativeElectricEnergyListDuplex struct {
	StartChannel   uint8
	ElectricEnergy []DirectionalEnergy
}

type InstantaneousElectricPowerListDuplex struct {
	StartChannel               uint8
	InstantaneousElectricPower []*int32
}

type DirectionalEnergy struct {
	NormalDirectionElectricEnergy  *uint32
	ReverseDirectionElectricEnergy *uint32
}

type CurrentPhases struct {
	RPhase *int16
	TPhase *int16
}

type LineVoltage struct {
	RS *uint16
	ST *uint16
}

type MeasurementChannel struct {
	ElectricEnergy *uint32
	CurrentRPhase  *int16
	CurrentTPhase  *int16
}

func decodeBoardCurrentList(edt []byte) (InstantaneousCurrentList, error) {
	if len(edt) < 2 {
		return InstantaneousCurrentList{}, fmt.Errorf("missing channel range")
	}
	start, count := edt[0], int(edt[1])
	if start == 0xFD && count == 0xFD && isNoData(edt[2:], 0x7FFE7FFE) {
		return InstantaneousCurrentList{}, nil
	}
	if start == 0 || count < 1 || count > 60 || int(start)+count-1 > 252 {
		return InstantaneousCurrentList{}, fmt.Errorf("invalid channel range: start %d, count %d", start, count)
	}
	if len(edt) != 2+4*count {
		return InstantaneousCurrentList{}, fmt.Errorf("expected %d bytes for %d channels, got %d", 2+4*count, count, len(edt))
	}
	values := make([]CurrentPhases, count)
	for i := range values {
		offset := 2 + 4*i
		current, err := decodeCurrentPhases(edt[offset : offset+4])
		if err != nil {
			return InstantaneousCurrentList{}, fmt.Errorf("channel %d: %w", int(start)+i, err)
		}
		values[i] = current
	}
	return InstantaneousCurrentList{StartChannel: start, InstantaneousCurrent: values}, nil
}

func decodeBoardEnergyListDuplex(edt []byte) (CumulativeElectricEnergyListDuplex, error) {
	if len(edt) < 2 {
		return CumulativeElectricEnergyListDuplex{}, fmt.Errorf("missing channel range")
	}
	start, count := edt[0], int(edt[1])
	if start == 0xFD && count == 0xFD && len(edt) == 10 &&
		isNoData(edt[2:6], noEnergyData) && isNoData(edt[6:10], noEnergyData) {
		return CumulativeElectricEnergyListDuplex{}, nil
	}
	if start == 0 || count < 1 || count > 30 || int(start)+count-1 > 252 {
		return CumulativeElectricEnergyListDuplex{}, fmt.Errorf("invalid channel range: start %d, count %d", start, count)
	}
	if len(edt) != 2+8*count {
		return CumulativeElectricEnergyListDuplex{}, fmt.Errorf("expected %d bytes for %d channels, got %d", 2+8*count, count, len(edt))
	}
	values := make([]DirectionalEnergy, count)
	for i := range values {
		current := &values[i]
		offset := 2 + 8*i
		for j, dest := range [2]**uint32{&current.NormalDirectionElectricEnergy, &current.ReverseDirectionElectricEnergy} {
			data := edt[offset+4*j : offset+4*(j+1)]
			if isNoData(data, noEnergyData) {
				continue
			}
			energy, err := decodeChannelEnergy(data)
			if err != nil {
				return CumulativeElectricEnergyListDuplex{}, fmt.Errorf("channel %d: %w", int(start)+i, err)
			}
			*dest = &energy
		}
	}
	return CumulativeElectricEnergyListDuplex{StartChannel: start, ElectricEnergy: values}, nil
}

func decodeCurrentPhases(edt []byte) (CurrentPhases, error) {
	if len(edt) != 4 {
		return CurrentPhases{}, fmt.Errorf("expected 4 bytes, got %d", len(edt))
	}
	var current CurrentPhases
	for i, dest := range [2]**int16{&current.RPhase, &current.TPhase} {
		raw := binary.BigEndian.Uint16(edt[2*i : 2*(i+1)])
		switch raw {
		case 0x8000, 0x7FFE, 0x7FFF: // Underflow, no data, overflow.
		default:
			*dest = new(int16(raw))
		}
	}
	return current, nil
}

func decodeLineVoltage(edt []byte) (LineVoltage, error) {
	if len(edt) != 4 {
		return LineVoltage{}, fmt.Errorf("expected 4 bytes, got %d", len(edt))
	}
	rs := binary.BigEndian.Uint16(edt[:2])
	st := binary.BigEndian.Uint16(edt[2:])
	var voltage LineVoltage
	if rs < 0xFFFE {
		voltage.RS = &rs
	}
	if st < 0xFFFE {
		voltage.ST = &st
	}
	return voltage, nil
}

func decodeMeasurementChannel(edt []byte) (MeasurementChannel, error) {
	if len(edt) != 8 {
		return MeasurementChannel{}, fmt.Errorf("expected 8 bytes, got %d", len(edt))
	}
	current, err := decodeCurrentPhases(edt[4:])
	if err != nil {
		return MeasurementChannel{}, err
	}
	channel := MeasurementChannel{CurrentRPhase: current.RPhase, CurrentTPhase: current.TPhase}
	if !isNoData(edt[:4], noEnergyData) {
		energy, err := decodeChannelEnergy(edt[:4])
		if err != nil {
			return MeasurementChannel{}, fmt.Errorf("electric energy: %w", err)
		}
		channel.ElectricEnergy = &energy
	}
	return channel, nil
}

package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
	"maps"
	"time"
)

const (
	EPCInstantaneousElectricPowerGeneration = 0xE0
	EPCCumulativeElectricEnergyOfGeneration = 0xE1
)

type PVPowerGenerationClient struct {
	conn *Connection
}

func NewPVPowerGenerationClient(conn *Connection) *PVPowerGenerationClient {
	return &PVPowerGenerationClient{
		conn: conn,
	}
}

func (p *PVPowerGenerationClient) Get(ctx context.Context, host string, eoj EOJ) (*PVPowerGeneration, error) {
	read := propertyReader{conn: p.conn, host: host, eoj: eoj, minimumWait: 20 * time.Second}
	values, err := read.required(ctx,
		EPCInstantaneousElectricPowerGeneration,
		EPCCumulativeElectricEnergyOfGeneration,
	)
	if err != nil {
		return nil, err
	}

	reading := &PVPowerGeneration{}
	reading.InstantaneousElectricPowerGeneration, err = decodeInstantaneousElectricPowerGeneration(values[EPCInstantaneousElectricPowerGeneration])
	if err != nil {
		return nil, fmt.Errorf("instantaneousElectricPowerGeneration (EPC 0x%02X): %w", EPCInstantaneousElectricPowerGeneration, err)
	}
	reading.CumulativeElectricEnergyOfGeneration, err = decodeMeterValue(values[EPCCumulativeElectricEnergyOfGeneration], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("cumulativeElectricEnergyOfGeneration (EPC 0x%02X): %w", EPCCumulativeElectricEnergyOfGeneration, err)
	}

	protocol, err := read.required(ctx, 0x82, epcGetPropertyMap)
	if err != nil {
		return nil, err
	}
	version := protocol[0x82]
	if len(version) != 4 {
		return nil, fmt.Errorf("invalid Appendix release (EPC 0x82): %X", version)
	}
	release := version[2]
	if release == 'a' { // Release A uses lowercase in EPC 0x82.
		release = 'A'
	}
	if release < 'A' || release > 'Z' {
		return nil, fmt.Errorf("invalid Appendix release (EPC 0x82): %X", version)
	}
	supported, err := decodePropertyMap(protocol[epcGetPropertyMap])
	if err != nil {
		return nil, fmt.Errorf("invalid Get property map: %w", err)
	}
	// PV AIF §3.2.1 groups status with the two generation EPCs read above.
	var status []uint8
	requiredIndividual := supportedEPCs(supported, 0xA0, 0xA1, 0xB4, 0xC3)
	optionalIndividual := supportedEPCs(supported, 0xE3, 0xE5, 0xE6, 0xE7, 0xE9)
	if release >= 'R' {
		status = append(status, EPCFaultDescription)
	}
	if release >= 'K' {
		status = append(status, 0xD0, 0xD1)
		requiredIndividual = append(requiredIndividual, 0xE8)
	} else {
		optionalIndividual = append(optionalIndividual, supportedEPCs(supported, 0xD0, 0xD1, 0xE8)...)
	}
	statusRequired, superOptional := superClassEPCs(supported, status...)
	monitoringValues, err := read.required(ctx, statusRequired...)
	if err != nil {
		return nil, err
	}
	for _, epc := range requiredIndividual {
		additional, err := read.required(ctx, epc)
		if err != nil {
			return nil, err
		}
		maps.Copy(monitoringValues, additional)
	}
	additional, err := read.optionalIndividually(ctx, append(superOptional, optionalIndividual...))
	if err != nil {
		return nil, err
	}
	maps.Copy(monitoringValues, additional)

	reading.SuperClass, err = decodeSuperClass(monitoringValues)
	if err != nil {
		return nil, err
	}

	for _, field := range []struct {
		epc      uint8
		min, max uint8
		dest     **uint8
	}{
		{0xA0, 0, 100, &reading.OutputPowerControl1},
		{0xD0, 0x00, 0x03, &reading.PowerSystemInterconnectionStatus},
		{0xD1, 0x41, 0x45, &reading.OutputPowerRestraintStatus},
		{0xE5, 0, 100, &reading.PowerGenerationOutputLimit1},
	} {
		edt := monitoringValues[field.epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeByte(edt, field.min, field.max)
		if err != nil {
			return nil, fmt.Errorf("EPC 0x%02X: %w", field.epc, err)
		}
		*field.dest = &value
	}
	for _, field := range []struct {
		epc       uint8
		max       uint16
		dest      **uint16
		noSetting bool
	}{
		{0xA1, 65533, &reading.OutputPowerControl2, false},
		{0xB4, 65533, &reading.UpperLimitClip, true},
		{0xC3, 65533, &reading.ApprovedCapacity, true},
		{0xE6, 65533, &reading.PowerGenerationOutputLimit2, false},
		{0xE7, 65533, &reading.LimitElectricEnergySold, false},
		{0xE8, 65533, &reading.RatedElectricPowerOfgeneration, true},
		{0xE9, 65533, &reading.RatedElectricPowerOfgenerationIndependent, false},
	} {
		edt := monitoringValues[field.epc]
		if len(edt) == 0 {
			continue
		}
		if len(edt) != 2 {
			return nil, fmt.Errorf("EPC 0x%02X: expected 2 bytes, got %d", field.epc, len(edt))
		}
		value := binary.BigEndian.Uint16(edt)
		if field.noSetting && value == 0xFFFF {
			continue
		}
		if value > field.max {
			return nil, fmt.Errorf("EPC 0x%02X: value %d exceeds %d", field.epc, value, field.max)
		}
		*field.dest = &value
	}
	if edt := monitoringValues[0xE3]; len(edt) != 0 {
		value, err := decodeMeterValue(edt, 999_999_999)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xE3: %w", err)
		}
		reading.CumulativeElectricEnergySold = &value
	}
	return reading, nil
}

type PVPowerGeneration struct {
	SuperClass

	// This property indicates instantaneous generated power in watts.
	InstantaneousElectricPowerGeneration uint16

	// This property indicates cumulative amounts of electric energy in 0.001 kWh.
	CumulativeElectricEnergyOfGeneration uint32

	OutputPowerControl1                       *uint8
	OutputPowerControl2                       *uint16
	UpperLimitClip                            *uint16
	ApprovedCapacity                          *uint16
	PowerSystemInterconnectionStatus          *uint8
	OutputPowerRestraintStatus                *uint8
	CumulativeElectricEnergySold              *uint32
	PowerGenerationOutputLimit1               *uint8
	PowerGenerationOutputLimit2               *uint16
	LimitElectricEnergySold                   *uint16
	RatedElectricPowerOfgeneration            *uint16
	RatedElectricPowerOfgenerationIndependent *uint16
}

func decodeInstantaneousElectricPowerGeneration(edt []byte) (uint16, error) {
	if len(edt) != 2 {
		return 0, fmt.Errorf("expected 2 bytes, got %d", len(edt))
	}
	value := binary.BigEndian.Uint16(edt)
	if value > 65_533 {
		return 0, fmt.Errorf("electric power out of range: %d", value)
	}
	return value, nil
}

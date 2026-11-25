package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
	"maps"
	"time"
)

const (
	EPCACChargeableElectricEnergy            = 0xA4
	EPCACDischargeableElectricEnergy         = 0xA5
	EPCACCumulativeChargingElectricEnergy    = 0xA8
	EPCACCumulativeDischargingElectricEnergy = 0xA9
)

type StorageBatteryClient struct {
	conn *Connection
}

func NewStorageBatteryClient(conn *Connection) *StorageBatteryClient {
	return &StorageBatteryClient{
		conn: conn,
	}
}

func (s *StorageBatteryClient) Get(ctx context.Context, host string, eoj EOJ) (*StorageBattery, error) {
	read := propertyReader{conn: s.conn, host: host, eoj: eoj, minimumWait: 20 * time.Second}
	protocol, err := read.required(ctx, 0x82, epcGetPropertyMap)
	if err != nil {
		return nil, err
	}
	version := protocol[0x82]
	if len(version) != 4 || version[2] < 'A' || version[2] > 'Z' {
		return nil, fmt.Errorf("invalid Appendix release (EPC 0x82): %X", version)
	}
	supported, err := decodePropertyMap(protocol[epcGetPropertyMap])
	if err != nil {
		return nil, fmt.Errorf("invalid Get property map: %w", err)
	}
	release := version[2]

	// Storage battery AIF §3.2.1, property group 1: status and remaining capacity.
	status := append([]uint8{0xDA}, supportedEPCs(supported, 0xE2, 0xE3, 0xE4)...)
	if release >= 'D' {
		status = append(status, 0xCF)
	}
	if release >= 'R' {
		status = append(status, EPCFaultDescription)
	}
	statusRequired, individual := superClassEPCs(supported, status...)
	individual = append(individual, supportedEPCs(supported, 0xA6, 0xA7,
		0xC7, 0xCA, 0xCB, 0xCC, 0xCD, 0xCE, 0xD3, 0xD4, 0xD5, 0xD6, 0xD8, 0xDC, 0xDD, 0xDE, 0xDF, 0xE5)...)
	if release < 'D' {
		individual = append(individual, supportedEPCs(supported, 0xCF)...)
	}
	monitoringValues, err := read.required(ctx, statusRequired...)
	if err != nil {
		return nil, err
	}

	// Storage battery AIF §3.2.1, property group 2: AC energy and interconnection.
	energy := []uint8{
		EPCACChargeableElectricEnergy,
		EPCACDischargeableElectricEnergy,
		EPCACCumulativeChargingElectricEnergy,
		EPCACCumulativeDischargingElectricEnergy,
	}
	// Storage battery AIF §3.1.4, property group 2: AC capacity and power limits.
	capacity := []uint8{0xA0, 0xA1, 0xA2, 0xA3, 0xC8, 0xC9}
	if release >= 'H' {
		energy = append(energy, 0xAA, 0xAB, 0xDB)
	} else {
		individual = append(individual, supportedEPCs(supported, 0xAA, 0xAB, 0xDB)...)
		individual = append(individual, supportedEPCs(supported, capacity...)...)
		capacity = nil
	}
	if release >= 'M' {
		capacity = append(capacity, 0xC1, 0xC2)
	} else {
		individual = append(individual, supportedEPCs(supported, 0xC1, 0xC2)...)
	}

	values, err := read.required(ctx, energy...)
	if err != nil {
		return nil, err
	}
	maps.Copy(monitoringValues, values)

	reading := &StorageBattery{}
	reading.ACChargeableElectricEnergy, err = decodeMeterValue(values[EPCACChargeableElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acChargeableElectricEnergy (EPC 0x%02X): %w", EPCACChargeableElectricEnergy, err)
	}
	reading.ACDischargeableElectricEnergy, err = decodeMeterValue(values[EPCACDischargeableElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acDischargeableElectricEnergy (EPC 0x%02X): %w", EPCACDischargeableElectricEnergy, err)
	}
	reading.ACCumulativeChargingElectricEnergy, err = decodeMeterValue(values[EPCACCumulativeChargingElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acCumulativeChargingElectricEnergy (EPC 0x%02X): %w", EPCACCumulativeChargingElectricEnergy, err)
	}
	reading.ACCumulativeDischargingElectricEnergy, err = decodeMeterValue(values[EPCACCumulativeDischargingElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acCumulativeDischargingElectricEnergy (EPC 0x%02X): %w", EPCACCumulativeDischargingElectricEnergy, err)
	}

	if len(capacity) != 0 {
		values, err := read.required(ctx, capacity...)
		if err != nil {
			return nil, err
		}
		maps.Copy(monitoringValues, values)
	}
	// Storage battery AIF §3.1.4, property group 1: ratings, if provided.
	ratings := supportedEPCs(supported, 0xD0, 0xD1, 0xD2)
	if len(ratings) != 0 {
		values, err := read.group(ctx, nil, ratings)
		if err != nil {
			return nil, err
		}
		maps.Copy(monitoringValues, values)
		// Retry optional EPCs omitted from a short Get_SNA.
		for _, epc := range ratings {
			if _, present := values[epc]; !present {
				individual = append(individual, epc)
			}
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
	}{
		{0xA6, 0, 100, &reading.ACChargeUpperLimit},
		{0xA7, 0, 100, &reading.ACDischargeLowerLimit},
		{0xC1, 0x00, 0x04, &reading.ChargingMethod},
		{0xC2, 0x00, 0x04, &reading.DischargingMethod},
		{0xCC, 0x41, 0x42, &reading.ReInterconnectionPermission},
		{0xCD, 0x41, 0x42, &reading.OperationPermission},
		{0xCE, 0x41, 0x42, &reading.IndependentOperationPermission},
		{0xDB, 0x00, 0x02, &reading.PowerSystemInterconnectionStatus},
		{0xE4, 0, 100, &reading.RemainingCapacity3},
		{0xE5, 0, 100, &reading.BatteryHealthState},
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
	if edt := monitoringValues[0xCF]; len(edt) != 0 {
		if len(edt) != 1 {
			return nil, fmt.Errorf("EPC 0xCF: expected 1 byte, got %d", len(edt))
		}
		switch edt[0] {
		case 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x48, 0x49:
			reading.ActualOperationMode = &edt[0]
		default:
			return nil, fmt.Errorf("EPC 0xCF: invalid state code 0x%02X", edt[0])
		}
	}
	if edt := monitoringValues[0xDA]; len(edt) != 0 {
		if len(edt) != 1 {
			return nil, fmt.Errorf("EPC 0xDA: expected 1 byte, got %d", len(edt))
		}
		switch edt[0] {
		case 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x48, 0x49:
			reading.OperationMode = &edt[0]
		default:
			return nil, fmt.Errorf("EPC 0xDA: invalid state code 0x%02X", edt[0])
		}
	}
	for _, field := range []struct {
		epc  uint8
		max  uint16
		dest **uint16
	}{
		{0xD1, 32766, &reading.RatedCapacity},
		{0xD2, 32766, &reading.RatedVoltage},
		{0xE3, 32766, &reading.RemainingCapacity2},
	} {
		edt := monitoringValues[field.epc]
		if len(edt) == 0 {
			continue
		}
		if len(edt) != 2 {
			return nil, fmt.Errorf("EPC 0x%02X: expected 2 bytes, got %d", field.epc, len(edt))
		}
		value := binary.BigEndian.Uint16(edt)
		if value > field.max {
			return nil, fmt.Errorf("EPC 0x%02X: value %d exceeds %d", field.epc, value, field.max)
		}
		*field.dest = &value
	}
	for _, field := range []struct {
		epc    uint8
		max    uint32
		dest   **uint32
		noData bool
	}{
		{0xA0, 999_999_999, &reading.ACEffectiveChargingCapacity, false},
		{0xA1, 999_999_999, &reading.ACEffectiveDischargingCapacity, false},
		{0xA2, 999_999_999, &reading.ACChargeableCapacity, false},
		{0xA3, 999_999_999, &reading.ACDischargeableCapacity, false},
		{0xAA, 999_999_999, &reading.ACTargetChargingElectricEnergy, true},
		{0xAB, 999_999_999, &reading.ACTargetDischargingElectricEnergy, true},
		{0xC7, 999_999_999, &reading.ACRatedElectricEnergy, false},
		{0xD0, 999_999_999, &reading.RatedElectricEnergy, false},
		{0xD6, 999_999_999, &reading.CumulativeDischargingElectricEnergy, false},
		{0xD8, 999_999_999, &reading.CumulativeChargingElectricEnergy, false},
		{0xE2, 999_999_999, &reading.RemainingCapacity1, false},
	} {
		edt := monitoringValues[field.epc]
		if len(edt) == 0 || (field.noData && isNoData(edt, 0)) {
			continue
		}
		value, err := decodeMeterValue(edt, field.max)
		if err != nil {
			return nil, fmt.Errorf("EPC 0x%02X: %w", field.epc, err)
		}
		*field.dest = &value
	}
	for _, field := range []struct {
		epc  uint8
		dest **int16
	}{
		{0xD4, &reading.InstantaneousChargingAndDischargingCurrent},
		{0xD5, &reading.InstantaneousChargingAndDischargingVoltage},
	} {
		edt := monitoringValues[field.epc]
		if len(edt) == 0 {
			continue
		}
		if len(edt) != 2 {
			return nil, fmt.Errorf("EPC 0x%02X: expected 2 bytes, got %d", field.epc, len(edt))
		}
		raw := binary.BigEndian.Uint16(edt)
		if raw == 0x8000 || raw == 0x7FFF {
			continue
		}
		*field.dest = new(int16(raw))
	}
	// The pair is encoded as the minimum followed by the maximum.
	for _, field := range []struct {
		epc  uint8
		dest *[2]*uint16
	}{
		{0xCA, &reading.MinimumAndMaximumChargingCurrent},
		{0xCB, &reading.MinimumAndMaximumDischargingCurrent},
		{0xDE, &reading.MinimumAndMaximumChargingCurrentAtIndependent},
		{0xDF, &reading.MinimumAndMaximumDischargingCurrentAtIndependent},
	} {
		edt := monitoringValues[field.epc]
		if len(edt) == 0 {
			continue
		}
		if len(edt) != 4 {
			return nil, fmt.Errorf("EPC 0x%02X: expected 4 bytes, got %d", field.epc, len(edt))
		}
		for i := range field.dest {
			value := binary.BigEndian.Uint16(edt[2*i : 2*(i+1)])
			if value == 0xFFFF {
				continue
			}
			if value > 32766 {
				return nil, fmt.Errorf("EPC 0x%02X: minimum or maximum out of range", field.epc)
			}
			field.dest[i] = &value
		}
	}
	// The pair is encoded as the minimum followed by the maximum.
	for _, field := range []struct {
		epc  uint8
		dest *[2]*uint32
	}{
		{0xC8, &reading.MinimumAndMaximumChargingElectricPower},
		{0xC9, &reading.MinimumAndMaximumDischargingElectricPower},
		{0xDC, &reading.MinimumAndMaximumChargingPowerAtIndependent},
		{0xDD, &reading.MinimumAndMaximumDischargingPowerAtIndependent},
	} {
		edt := monitoringValues[field.epc]
		if len(edt) == 0 {
			continue
		}
		if len(edt) != 8 {
			return nil, fmt.Errorf("EPC 0x%02X: expected 8 bytes, got %d", field.epc, len(edt))
		}
		for i := range field.dest {
			value := binary.BigEndian.Uint32(edt[4*i : 4*(i+1)])
			if value == 0xFFFFFFFF {
				continue
			}
			if value > 999_999_999 {
				return nil, fmt.Errorf("EPC 0x%02X: minimum or maximum out of range", field.epc)
			}
			field.dest[i] = &value
		}
	}
	if edt := monitoringValues[0xD3]; len(edt) != 0 {
		if len(edt) != 4 {
			return nil, fmt.Errorf("EPC 0xD3: expected 4 bytes, got %d", len(edt))
		}
		raw := binary.BigEndian.Uint32(edt)
		if raw != 0x80000000 && raw != 0x7FFFFFFF {
			value := int32(raw)
			if value < -999_999_999 || value > 999_999_999 {
				return nil, fmt.Errorf("EPC 0xD3: electric power out of range: %d", value)
			}
			reading.InstantaneousChargingAndDischargingElectricPower = &value
		}
	}
	if reading.RemainingCapacity1 == nil && reading.RemainingCapacity2 == nil && reading.RemainingCapacity3 == nil {
		return nil, fmt.Errorf("Get property map lacks a remaining capacity property (EPC 0xE2, 0xE3, or 0xE4)")
	}
	return reading, nil
}

type StorageBattery struct {
	SuperClass

	// This property indicates the electric energy that can be charged at the
	// present point in time (AC).
	ACChargeableElectricEnergy uint32

	// This property indicates the electric energy that can be discharged at the
	// present point in time (AC).
	ACDischargeableElectricEnergy uint32

	// This property indicates the cumulative charging electric energy (AC) in 0.001kWh.
	ACCumulativeChargingElectricEnergy uint32

	// This property indicates the cumulative discharging electric energy (AC) in 0.001kWh.
	ACCumulativeDischargingElectricEnergy uint32

	ACEffectiveChargingCapacity                      *uint32
	ACEffectiveDischargingCapacity                   *uint32
	ACChargeableCapacity                             *uint32
	ACDischargeableCapacity                          *uint32
	ACChargeUpperLimit                               *uint8
	ACDischargeLowerLimit                            *uint8
	ACTargetChargingElectricEnergy                   *uint32
	ACTargetDischargingElectricEnergy                *uint32
	ChargingMethod                                   *uint8
	DischargingMethod                                *uint8
	ACRatedElectricEnergy                            *uint32
	MinimumAndMaximumChargingElectricPower           [2]*uint32
	MinimumAndMaximumDischargingElectricPower        [2]*uint32
	MinimumAndMaximumChargingCurrent                 [2]*uint16
	MinimumAndMaximumDischargingCurrent              [2]*uint16
	ReInterconnectionPermission                      *uint8
	OperationPermission                              *uint8
	IndependentOperationPermission                   *uint8
	ActualOperationMode                              *uint8
	RatedElectricEnergy                              *uint32
	RatedCapacity                                    *uint16
	RatedVoltage                                     *uint16
	InstantaneousChargingAndDischargingElectricPower *int32
	InstantaneousChargingAndDischargingCurrent       *int16
	InstantaneousChargingAndDischargingVoltage       *int16
	CumulativeDischargingElectricEnergy              *uint32
	CumulativeChargingElectricEnergy                 *uint32
	OperationMode                                    *uint8
	PowerSystemInterconnectionStatus                 *uint8
	MinimumAndMaximumChargingPowerAtIndependent      [2]*uint32
	MinimumAndMaximumDischargingPowerAtIndependent   [2]*uint32
	MinimumAndMaximumChargingCurrentAtIndependent    [2]*uint16
	MinimumAndMaximumDischargingCurrentAtIndependent [2]*uint16
	RemainingCapacity1                               *uint32
	RemainingCapacity2                               *uint16
	RemainingCapacity3                               *uint8
	BatteryHealthState                               *uint8
}

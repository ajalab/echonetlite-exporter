package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
	"maps"
	"time"
)

const (
	EPCHotWaterHeatingStatus          = 0xD0
	EPCTargetSuppliedWaterTemperature = 0xD1
	EPCTargetBathWaterTemperature     = 0xE1
	EPCBathWaterHeatingStatus         = 0xE2
	EPCAutomaticBathOperation         = 0xE3
	EPCShowerHotWaterSupplyStatus     = 0xEA
	EPCKitchenHotWaterSupplyStatus    = 0xEB
	EPCBathOperationStatus            = 0xEF
)

type InstantaneousWaterHeater struct {
	SuperClass

	HotWaterHeatingStatus          uint8
	BathWaterHeatingStatus         uint8
	AutomaticBathOperation         *uint8
	TargetSuppliedWaterTemperature *uint8 // °C
	TargetBathWaterTemperature     *uint8 // °C
	ShowerHotWaterSupplyStatus     *uint8
	KitchenHotWaterSupplyStatus    *uint8
	BathOperationStatus            *uint8

	HotWaterWarmerSetting               *uint8
	RemainingAutomaticOperationTime     *uint32
	TargetBathAdditionalBoilupOperation *uint8
	BathHotWaterAddition                *uint8
	BathLukewarmWaterFunction           *uint8
	BathWaterVolume1                    *uint8
	BathroomPriority                    *uint8
	BathWaterVolume3                    *uint16
}

type InstantaneousWaterHeaterClient struct {
	conn *Connection
}

func NewInstantaneousWaterHeaterClient(conn *Connection) *InstantaneousWaterHeaterClient {
	return &InstantaneousWaterHeaterClient{conn: conn}
}

func (c *InstantaneousWaterHeaterClient) Get(ctx context.Context, host string, eoj EOJ) (*InstantaneousWaterHeater, error) {
	read := propertyReader{conn: c.conn, host: host, eoj: eoj, minimumWait: 10 * time.Second}
	supported, err := read.readableProperties(ctx)
	if err != nil {
		return nil, err
	}

	// AIF §3.1.4 allows these device properties to be read together.
	values, err := read.group(ctx, []uint8{EPCHotWaterHeatingStatus, EPCBathWaterHeatingStatus},
		supportedEPCs(supported, EPCAutomaticBathOperation))
	if err != nil {
		return nil, err
	}
	hotWaterHeating, err := decodeByte(values[EPCHotWaterHeatingStatus], 0x41, 0x42)
	if err != nil {
		return nil, fmt.Errorf("hotWaterHeatingStatus (EPC 0xD0): %w", err)
	}
	bathWaterHeating, err := decodeByte(values[EPCBathWaterHeatingStatus], 0x41, 0x42)
	if err != nil {
		return nil, fmt.Errorf("bathWaterHeatingStatus (EPC 0xE2): %w", err)
	}

	optional, err := read.optionalIndividually(ctx, supportedEPCs(supported,
		EPCTargetSuppliedWaterTemperature, EPCTargetBathWaterTemperature,
		EPCShowerHotWaterSupplyStatus, EPCKitchenHotWaterSupplyStatus, EPCBathOperationStatus))
	if err != nil {
		return nil, err
	}
	maps.Copy(values, optional)
	reading := &InstantaneousWaterHeater{
		HotWaterHeatingStatus:  hotWaterHeating,
		BathWaterHeatingStatus: bathWaterHeating,
	}
	for _, field := range []struct {
		epc  uint8
		dest **uint8
	}{
		{EPCAutomaticBathOperation, &reading.AutomaticBathOperation},
		{EPCShowerHotWaterSupplyStatus, &reading.ShowerHotWaterSupplyStatus},
		{EPCKitchenHotWaterSupplyStatus, &reading.KitchenHotWaterSupplyStatus},
	} {
		edt := values[field.epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeByte(edt, 0x41, 0x42)
		if err != nil {
			return nil, fmt.Errorf("heating state (EPC 0x%02X): %w", field.epc, err)
		}
		*field.dest = &value
	}
	for _, field := range []struct {
		epc      uint8
		min, max uint8
		dest     **uint8
	}{
		{EPCTargetSuppliedWaterTemperature, 0, 100, &reading.TargetSuppliedWaterTemperature},
		{EPCTargetBathWaterTemperature, 0, 100, &reading.TargetBathWaterTemperature},
		{EPCBathOperationStatus, 0x41, 0x43, &reading.BathOperationStatus},
	} {
		edt := values[field.epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeByte(edt, field.min, field.max)
		if err != nil {
			return nil, fmt.Errorf("water heater property (EPC 0x%02X): %w", field.epc, err)
		}
		*field.dest = &value
	}

	monitoringRequired, monitoringOptional := superClassEPCs(supported)
	monitoringValues, err := read.group(ctx, monitoringRequired, monitoringOptional)
	if err != nil {
		return nil, err
	}
	individual := supportedEPCs(supported, 0xD2, 0xDB, 0xE4, 0xE5, 0xE6, 0xE7, 0xE9, 0xEE)
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
	}{
		{0xD2, 0x41, 0x42, &reading.HotWaterWarmerSetting},
		{0xE4, 0x41, 0x42, &reading.TargetBathAdditionalBoilupOperation},
		{0xE5, 0x41, 0x42, &reading.BathHotWaterAddition},
		{0xE6, 0x41, 0x42, &reading.BathLukewarmWaterFunction},
		{0xE9, 0x41, 0x42, &reading.BathroomPriority},
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
	// The two highest codes in each bath volume property denote underflow and overflow.
	if edt := monitoringValues[0xE7]; len(edt) != 0 {
		if len(edt) != 1 {
			return nil, fmt.Errorf("EPC 0xE7: expected 1 byte, got %d", len(edt))
		}
		if edt[0] < 0xFE {
			reading.BathWaterVolume1 = &edt[0]
		}
	}
	if edt := monitoringValues[0xEE]; len(edt) != 0 {
		if len(edt) != 2 {
			return nil, fmt.Errorf("EPC 0xEE: expected 2 bytes, got %d", len(edt))
		}
		value := binary.BigEndian.Uint16(edt)
		if value < 0xFFFE {
			reading.BathWaterVolume3 = &value
		}
	}
	if edt := monitoringValues[0xDB]; len(edt) != 0 {
		switch {
		case len(edt) == 2 && edt[0] == 0xFF && edt[1] == 0xFF:
			// The device reports unlimited remaining time.
		case len(edt) == 2 && edt[0] <= 23 && edt[1] <= 59:
			reading.RemainingAutomaticOperationTime = new(uint32(edt[0])*3600 + uint32(edt[1])*60)
		default:
			return nil, fmt.Errorf("EPC 0xDB: invalid remaining automatic operation time: %X", edt)
		}
	}
	return reading, nil
}

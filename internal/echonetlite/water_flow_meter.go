package echonetlite

import (
	"context"
	"fmt"
	"maps"
	"math"
)

const (
	EPCCumulativeFlowingWater        = 0xE0
	epcUnitForCumulativeFlowingWater = 0xE1
)

type WaterFlowMeter struct {
	SuperClass

	CumulativeFlowingWater        uint32
	UnitForCumulativeFlowingWater float64 // m³ per count

	AbnormalDetection       *uint8
	SecurityDataInformation *uint32
}

type WaterFlowMeterClient struct {
	conn *Connection
}

func NewWaterFlowMeterClient(conn *Connection) *WaterFlowMeterClient {
	return &WaterFlowMeterClient{conn: conn}
}

func (c *WaterFlowMeterClient) Get(ctx context.Context, host string, eoj EOJ) (*WaterFlowMeter, error) {
	read := propertyReader{conn: c.conn, host: host, eoj: eoj}
	values, err := read.required(ctx, EPCCumulativeFlowingWater, epcUnitForCumulativeFlowingWater)
	if err != nil {
		return nil, err
	}
	volume, err := decodeMeterValue(values[EPCCumulativeFlowingWater], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("invalid cumulativeFlowingWater: %w", err)
	}
	unit, err := decodeByte(values[epcUnitForCumulativeFlowingWater], 0x00, 0x06)
	if err != nil {
		return nil, fmt.Errorf("unitForCumulativeFlowingWater (EPC 0xE1): %w", err)
	}
	reading := &WaterFlowMeter{
		CumulativeFlowingWater:        volume,
		UnitForCumulativeFlowingWater: math.Pow10(-int(unit)),
	}

	supported, err := read.readableProperties(ctx)
	if err != nil {
		return nil, err
	}
	monitoringRequired, monitoringOptional := superClassEPCs(supported)
	monitoringOptional = append(monitoringOptional, supportedEPCs(supported, 0xE3, 0xE4)...)
	monitoringValues, err := read.group(ctx, monitoringRequired, monitoringOptional)
	if err != nil {
		return nil, err
	}
	var individual []uint8
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

	if edt := monitoringValues[0xE3]; len(edt) != 0 {
		value, err := decodeByte(edt, 0x41, 0x42)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xE3: %w", err)
		}
		reading.AbnormalDetection = &value
	}
	if edt := monitoringValues[0xE4]; len(edt) != 0 {
		value, err := decodeMeterValue(edt, 0xFFFFFFFF)
		if err != nil {
			return nil, fmt.Errorf("EPC 0xE4: %w", err)
		}
		reading.SecurityDataInformation = &value
	}
	return reading, nil
}

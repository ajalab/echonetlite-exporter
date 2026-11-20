package echonetlite

import (
	"context"
	"fmt"
	"math"
)

const (
	epcCumulativeFlowingWater        = 0xE0
	epcUnitForCumulativeFlowingWater = 0xE1
)

type WaterFlowMeter struct {
	CumulativeFlowingWater        uint32
	UnitForCumulativeFlowingWater float64 // m³ per count
}

type WaterFlowMeterClient struct {
	conn *Connection
}

func NewWaterFlowMeterClient(conn *Connection) *WaterFlowMeterClient {
	return &WaterFlowMeterClient{conn: conn}
}

func (c *WaterFlowMeterClient) Get(ctx context.Context, host string, eoj EOJ) (*WaterFlowMeter, error) {
	read := propertyReader{conn: c.conn, host: host, eoj: eoj}
	values, err := read.required(ctx, epcCumulativeFlowingWater, epcUnitForCumulativeFlowingWater)
	if err != nil {
		return nil, err
	}
	volume, err := decodeMeterValue(values[epcCumulativeFlowingWater], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("invalid cumulativeFlowingWater: %w", err)
	}
	unit, err := decodeByte(values[epcUnitForCumulativeFlowingWater], 0x00, 0x06)
	if err != nil {
		return nil, fmt.Errorf("unitForCumulativeFlowingWater (EPC 0xE1): %w", err)
	}
	return &WaterFlowMeter{
		CumulativeFlowingWater:        volume,
		UnitForCumulativeFlowingWater: math.Pow10(-int(unit)),
	}, nil
}

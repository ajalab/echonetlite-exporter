package echonetlite

import (
	"context"
	"fmt"
	"math"
)

const (
	EPCCumulativeElectricEnergy                  = 0xE0
	epcWattHourMeterCumulativeElectricEnergyUnit = 0xE2
)

type WattHourMeter struct {
	CumulativeElectricEnergy              uint32
	CumulativeAmountsOfElectricEnergyUnit float64 // kWh per count
}

type WattHourMeterClient struct {
	conn *Connection
}

func NewWattHourMeterClient(conn *Connection) *WattHourMeterClient {
	return &WattHourMeterClient{conn: conn}
}

func (c *WattHourMeterClient) Get(ctx context.Context, host string, eoj EOJ) (*WattHourMeter, error) {
	read := propertyReader{conn: c.conn, host: host, eoj: eoj}
	values, err := read.required(ctx, EPCCumulativeElectricEnergy, epcWattHourMeterCumulativeElectricEnergyUnit)
	if err != nil {
		return nil, err
	}
	energy, err := decodeMeterValue(values[EPCCumulativeElectricEnergy], 99_999_999)
	if err != nil {
		return nil, fmt.Errorf("invalid cumulativeElectricEnergy: %w", err)
	}
	unit, err := decodeByte(values[epcWattHourMeterCumulativeElectricEnergyUnit], 0x01, 0x02)
	if err != nil {
		return nil, fmt.Errorf("cumulativeAmountsOfElectricEnergyUnit (EPC 0xE2): %w", err)
	}
	return &WattHourMeter{
		CumulativeElectricEnergy:              energy,
		CumulativeAmountsOfElectricEnergyUnit: math.Pow10(-int(unit)),
	}, nil
}

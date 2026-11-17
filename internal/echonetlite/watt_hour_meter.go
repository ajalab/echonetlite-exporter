package echonetlite

import (
	"context"
	"fmt"
	"math"
)

const (
	epcCumulativeElectricEnergy                  = 0xE0
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
	values, err := getRequiredProperties(ctx, c.conn, host, eoj,
		epcCumulativeElectricEnergy, epcWattHourMeterCumulativeElectricEnergyUnit)
	if err != nil {
		return nil, err
	}
	energy, err := decodeMeterValue(values[epcCumulativeElectricEnergy], 99_999_999)
	if err != nil {
		return nil, fmt.Errorf("invalid cumulativeElectricEnergy: %w", err)
	}
	unit := decodeByte(values[epcWattHourMeterCumulativeElectricEnergyUnit], 0x01, 0x02)
	if unit == nil {
		return nil, fmt.Errorf("invalid cumulativeAmountsOfElectricEnergyUnit: %X", values[epcWattHourMeterCumulativeElectricEnergyUnit])
	}
	return &WattHourMeter{
		CumulativeElectricEnergy:              energy,
		CumulativeAmountsOfElectricEnergyUnit: math.Pow10(-int(*unit)),
	}, nil
}

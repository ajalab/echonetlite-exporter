package echonetlite

import (
	"context"
	"fmt"
	"maps"
	"math"
)

const (
	EPCCumulativeElectricEnergy                  = 0xE0
	epcWattHourMeterCumulativeElectricEnergyUnit = 0xE2
)

type WattHourMeter struct {
	SuperClass

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
	reading := &WattHourMeter{
		CumulativeElectricEnergy:              energy,
		CumulativeAmountsOfElectricEnergyUnit: math.Pow10(-int(unit)),
	}

	supported, err := read.readableProperties(ctx)
	if err != nil {
		return nil, err
	}
	monitoringRequired, monitoringOptional := superClassEPCs(supported)
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

	return reading, nil
}

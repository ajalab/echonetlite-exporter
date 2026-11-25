package echonetlite

import (
	"context"
	"fmt"
	"maps"
)

const EPCCumulativeGasConsumption = 0xE0

type GasMeter struct {
	SuperClass

	CumulativeGasConsumption uint32 // in units of 0.001 m³
}

type GasMeterClient struct {
	conn *Connection
}

func NewGasMeterClient(conn *Connection) *GasMeterClient {
	return &GasMeterClient{conn: conn}
}

func (c *GasMeterClient) Get(ctx context.Context, host string, eoj EOJ) (*GasMeter, error) {
	read := propertyReader{conn: c.conn, host: host, eoj: eoj}
	values, err := read.required(ctx, EPCCumulativeGasConsumption)
	if err != nil {
		return nil, err
	}
	consumption, err := decodeMeterValue(values[EPCCumulativeGasConsumption], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("invalid cumulativeGasConsumption: %w", err)
	}
	reading := &GasMeter{CumulativeGasConsumption: consumption}

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

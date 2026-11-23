package echonetlite

import (
	"context"
	"fmt"
)

const EPCCumulativeGasConsumption = 0xE0

type GasMeter struct {
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
	return &GasMeter{CumulativeGasConsumption: consumption}, nil
}

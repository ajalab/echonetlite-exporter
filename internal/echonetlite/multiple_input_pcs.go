package echonetlite

import (
	"context"
	"fmt"
)

const (
	EPCMultipleInputPCSNormalDirectionElectricEnergy  = 0xE0
	EPCMultipleInputPCSReverseDirectionElectricEnergy = 0xE3
	EPCMultipleInputPCSInstantaneousElectricPower     = 0xE7
)

type MultipleInputPCSClient struct {
	conn *Connection
}

func NewMultipleInputPCSClient(conn *Connection) *MultipleInputPCSClient {
	return &MultipleInputPCSClient{conn: conn}
}

func (c *MultipleInputPCSClient) Get(ctx context.Context, host string, eoj EOJ) (*MultipleInputPCS, error) {
	read := propertyReader{conn: c.conn, host: host, eoj: eoj}
	values, err := read.required(ctx,
		EPCMultipleInputPCSNormalDirectionElectricEnergy,
		EPCMultipleInputPCSReverseDirectionElectricEnergy,
		EPCMultipleInputPCSInstantaneousElectricPower,
	)
	if err != nil {
		return nil, err
	}

	reading := &MultipleInputPCS{}
	reading.NormalDirectionElectricEnergy, err = decodeMeterValue(values[EPCMultipleInputPCSNormalDirectionElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("normalDirectionElectricEnergy (EPC 0x%02X): %w", EPCMultipleInputPCSNormalDirectionElectricEnergy, err)
	}
	reading.ReverseDirectionElectricEnergy, err = decodeMeterValue(values[EPCMultipleInputPCSReverseDirectionElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("reverseDirectionElectricEnergy (EPC 0x%02X): %w", EPCMultipleInputPCSReverseDirectionElectricEnergy, err)
	}
	reading.InstantaneousElectricPower, err = decodeElectricPower(values[EPCMultipleInputPCSInstantaneousElectricPower])
	if err != nil {
		return nil, fmt.Errorf("instantaneousElectricPower (EPC 0x%02X): %w", EPCMultipleInputPCSInstantaneousElectricPower, err)
	}
	return reading, nil
}

type MultipleInputPCS struct {
	// This property indicates the cumulative amount of electric energy in the
	// normal direction (AC-DC conversion) in 0.001 kWh.
	NormalDirectionElectricEnergy uint32

	// This property indicates the cumulative amount of electric energy in the
	// reverse direction (DC-AC conversion) in 0.001 kWh.
	ReverseDirectionElectricEnergy uint32

	// This property indicates the measured instantaneous electric power in watts.
	InstantaneousElectricPower int32
}

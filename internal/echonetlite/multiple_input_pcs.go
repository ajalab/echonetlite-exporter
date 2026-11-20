package echonetlite

import (
	"context"
	"fmt"
)

const (
	epcMultipleInputPCSNormalDirectionElectricEnergy  = 0xE0
	epcMultipleInputPCSReverseDirectionElectricEnergy = 0xE3
	epcMultipleInputPCSInstantaneousElectricPower     = 0xE7
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
		epcMultipleInputPCSNormalDirectionElectricEnergy,
		epcMultipleInputPCSReverseDirectionElectricEnergy,
		epcMultipleInputPCSInstantaneousElectricPower,
	)
	if err != nil {
		return nil, err
	}

	reading := &MultipleInputPCS{}
	reading.NormalDirectionElectricEnergy, err = decodeMeterValue(values[epcMultipleInputPCSNormalDirectionElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("normalDirectionElectricEnergy (EPC 0x%02X): %w", epcMultipleInputPCSNormalDirectionElectricEnergy, err)
	}
	reading.ReverseDirectionElectricEnergy, err = decodeMeterValue(values[epcMultipleInputPCSReverseDirectionElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("reverseDirectionElectricEnergy (EPC 0x%02X): %w", epcMultipleInputPCSReverseDirectionElectricEnergy, err)
	}
	reading.InstantaneousElectricPower, err = decodeElectricPower(values[epcMultipleInputPCSInstantaneousElectricPower])
	if err != nil {
		return nil, fmt.Errorf("instantaneousElectricPower (EPC 0x%02X): %w", epcMultipleInputPCSInstantaneousElectricPower, err)
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

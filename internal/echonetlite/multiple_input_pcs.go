package echonetlite

import (
	"context"
	"encoding/binary"
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
	getReq := Frame{
		SEOJ: EOJ{0x05, 0xFF, 0x01},
		DEOJ: eoj,
		ESV:  0x62,
		Properties: []Property{
			{EPC: epcMultipleInputPCSNormalDirectionElectricEnergy, EDT: []byte{}},
			{EPC: epcMultipleInputPCSReverseDirectionElectricEnergy, EDT: []byte{}},
			{EPC: epcMultipleInputPCSInstantaneousElectricPower, EDT: []byte{}},
		},
	}
	resFrame, err := c.conn.Unicast(ctx, host, getReq)
	if err != nil {
		return nil, err
	}
	if len(resFrame.Properties) != len(getReq.Properties) {
		return nil, fmt.Errorf("unexpected property count: got %d, want %d", len(resFrame.Properties), len(getReq.Properties))
	}

	mipcs := &MultipleInputPCS{}
	for _, prop := range resFrame.Properties {
		switch prop.EPC {
		case epcMultipleInputPCSNormalDirectionElectricEnergy:
			val, err := parseMultipleInputPCSElectricEnergy(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid normalDirectionElectricEnergy data: %w", err)
			}
			mipcs.NormalDirectionElectricEnergy = val
		case epcMultipleInputPCSReverseDirectionElectricEnergy:
			val, err := parseMultipleInputPCSElectricEnergy(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid reverseDirectionElectricEnergy data: %w", err)
			}
			mipcs.ReverseDirectionElectricEnergy = val
		case epcMultipleInputPCSInstantaneousElectricPower:
			val, err := parseMultipleInputPCSInstantaneousElectricPower(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid instantaneousElectricPower data: %w", err)
			}
			mipcs.InstantaneousElectricPower = val
		default:
			return nil, fmt.Errorf("unexpected EPC: 0x%X", prop.EPC)
		}
	}

	return mipcs, nil
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

func parseMultipleInputPCSElectricEnergy(edts []byte) (uint32, error) {
	if len(edts) != 4 {
		return 0, fmt.Errorf("expected 4 bytes, got %d", len(edts))
	}
	return binary.BigEndian.Uint32(edts), nil
}

func parseMultipleInputPCSInstantaneousElectricPower(edts []byte) (int32, error) {
	if len(edts) != 4 {
		return 0, fmt.Errorf("expected 4 bytes, got %d", len(edts))
	}
	return int32(binary.BigEndian.Uint32(edts)), nil
}

package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
)

const (
	epcACChargeableElectricEnergy    = 0xA4
	epcACDischargeableElectricEnergy = 0xA5
)

type StorageBatteryClient struct {
	conn *Connection
}

func NewStorageBatteryClient(conn *Connection) *StorageBatteryClient {
	return &StorageBatteryClient{
		conn: conn,
	}
}

func (s *StorageBatteryClient) Get(ctx context.Context, host string, eoj EOJ) (*StorageBattery, error) {
	getReq := Frame{
		SEOJ: EOJ{0x05, 0xFF, 0x01},
		DEOJ: eoj,
		ESV:  0x62,
		Properties: []Property{
			{EPC: epcACChargeableElectricEnergy, EDT: []byte{}},
			{EPC: epcACDischargeableElectricEnergy, EDT: []byte{}},
		},
	}
	resFrame, err := s.conn.Unicast(ctx, host, getReq)
	if err != nil {
		return nil, err
	}
	if len(resFrame.Properties) != len(getReq.Properties) {
		return nil, fmt.Errorf("unexpected property count: got %d, want %d", len(resFrame.Properties), len(getReq.Properties))
	}

	sb := &StorageBattery{}
	for _, prop := range resFrame.Properties {
		switch prop.EPC {
		case epcACChargeableElectricEnergy:
			val, err := parseStorageBatteryWh(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid acChargeableElectricEnergy data: %w", err)
			}
			sb.ACChargeableElectricEnergyWh = val
		case epcACDischargeableElectricEnergy:
			val, err := parseStorageBatteryWh(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid acDischargeableElectricEnergy data: %w", err)
			}
			sb.ACDischargeableElectricEnergyWh = val
		default:
			return nil, fmt.Errorf("unexpected EPC: 0x%X", prop.EPC)
		}
	}

	return sb, nil
}

type StorageBattery struct {
	// This property indicates the electric energy that can be charged at the
	// present point in time (AC).
	ACChargeableElectricEnergyWh uint32

	// This property indicates the electric energy that can be discharged at the
	// present point in time (AC).
	ACDischargeableElectricEnergyWh uint32
}

func parseStorageBatteryWh(edts []byte) (uint32, error) {
	if len(edts) != 4 {
		return 0, fmt.Errorf("expected 4 bytes, got %d", len(edts))
	}
	return binary.BigEndian.Uint32(edts), nil
}

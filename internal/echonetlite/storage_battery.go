package echonetlite

import (
	"context"
	"fmt"
	"time"
)

const (
	epcACChargeableElectricEnergy            = 0xA4
	epcACDischargeableElectricEnergy         = 0xA5
	epcACCumulativeChargingElectricEnergy    = 0xA8
	epcACCumulativeDischargingElectricEnergy = 0xA9
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
	read := propertyReader{conn: s.conn, host: host, eoj: eoj, minimumWait: 20 * time.Second}
	values, err := read.required(ctx,
		epcACChargeableElectricEnergy,
		epcACDischargeableElectricEnergy,
		epcACCumulativeChargingElectricEnergy,
		epcACCumulativeDischargingElectricEnergy,
	)
	if err != nil {
		return nil, err
	}

	reading := &StorageBattery{}
	reading.ACChargeableElectricEnergy, err = decodeMeterValue(values[epcACChargeableElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acChargeableElectricEnergy (EPC 0x%02X): %w", epcACChargeableElectricEnergy, err)
	}
	reading.ACDischargeableElectricEnergy, err = decodeMeterValue(values[epcACDischargeableElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acDischargeableElectricEnergy (EPC 0x%02X): %w", epcACDischargeableElectricEnergy, err)
	}
	reading.ACCumulativeChargingElectricEnergy, err = decodeMeterValue(values[epcACCumulativeChargingElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acCumulativeChargingElectricEnergy (EPC 0x%02X): %w", epcACCumulativeChargingElectricEnergy, err)
	}
	reading.ACCumulativeDischargingElectricEnergy, err = decodeMeterValue(values[epcACCumulativeDischargingElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acCumulativeDischargingElectricEnergy (EPC 0x%02X): %w", epcACCumulativeDischargingElectricEnergy, err)
	}
	return reading, nil
}

type StorageBattery struct {
	// This property indicates the electric energy that can be charged at the
	// present point in time (AC).
	ACChargeableElectricEnergy uint32

	// This property indicates the electric energy that can be discharged at the
	// present point in time (AC).
	ACDischargeableElectricEnergy uint32

	// This property indicates the cumulative charging electric energy (AC) in 0.001kWh.
	ACCumulativeChargingElectricEnergy uint32

	// This property indicates the cumulative discharging electric energy (AC) in 0.001kWh.
	ACCumulativeDischargingElectricEnergy uint32
}

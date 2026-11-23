package echonetlite

import (
	"context"
	"fmt"
	"time"
)

const (
	EPCACChargeableElectricEnergy            = 0xA4
	EPCACDischargeableElectricEnergy         = 0xA5
	EPCACCumulativeChargingElectricEnergy    = 0xA8
	EPCACCumulativeDischargingElectricEnergy = 0xA9
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
		EPCACChargeableElectricEnergy,
		EPCACDischargeableElectricEnergy,
		EPCACCumulativeChargingElectricEnergy,
		EPCACCumulativeDischargingElectricEnergy,
	)
	if err != nil {
		return nil, err
	}

	reading := &StorageBattery{}
	reading.ACChargeableElectricEnergy, err = decodeMeterValue(values[EPCACChargeableElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acChargeableElectricEnergy (EPC 0x%02X): %w", EPCACChargeableElectricEnergy, err)
	}
	reading.ACDischargeableElectricEnergy, err = decodeMeterValue(values[EPCACDischargeableElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acDischargeableElectricEnergy (EPC 0x%02X): %w", EPCACDischargeableElectricEnergy, err)
	}
	reading.ACCumulativeChargingElectricEnergy, err = decodeMeterValue(values[EPCACCumulativeChargingElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acCumulativeChargingElectricEnergy (EPC 0x%02X): %w", EPCACCumulativeChargingElectricEnergy, err)
	}
	reading.ACCumulativeDischargingElectricEnergy, err = decodeMeterValue(values[EPCACCumulativeDischargingElectricEnergy], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("acCumulativeDischargingElectricEnergy (EPC 0x%02X): %w", EPCACCumulativeDischargingElectricEnergy, err)
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

package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	EPCInstantaneousElectricPowerGeneration = 0xE0
	EPCCumulativeElectricEnergyOfGeneration = 0xE1
)

type PVPowerGenerationClient struct {
	conn *Connection
}

func NewPVPowerGenerationClient(conn *Connection) *PVPowerGenerationClient {
	return &PVPowerGenerationClient{
		conn: conn,
	}
}

func (p *PVPowerGenerationClient) Get(ctx context.Context, host string, eoj EOJ) (*PVPowerGeneration, error) {
	read := propertyReader{conn: p.conn, host: host, eoj: eoj, minimumWait: 20 * time.Second}
	values, err := read.required(ctx,
		EPCInstantaneousElectricPowerGeneration,
		EPCCumulativeElectricEnergyOfGeneration,
	)
	if err != nil {
		return nil, err
	}

	reading := &PVPowerGeneration{}
	reading.InstantaneousElectricPowerGeneration, err = decodeInstantaneousElectricPowerGeneration(values[EPCInstantaneousElectricPowerGeneration])
	if err != nil {
		return nil, fmt.Errorf("instantaneousElectricPowerGeneration (EPC 0x%02X): %w", EPCInstantaneousElectricPowerGeneration, err)
	}
	reading.CumulativeElectricEnergyOfGeneration, err = decodeMeterValue(values[EPCCumulativeElectricEnergyOfGeneration], 999_999_999)
	if err != nil {
		return nil, fmt.Errorf("cumulativeElectricEnergyOfGeneration (EPC 0x%02X): %w", EPCCumulativeElectricEnergyOfGeneration, err)
	}
	return reading, nil
}

type PVPowerGeneration struct {
	// This property indicates instantaneous generated power in watts.
	InstantaneousElectricPowerGeneration uint16

	// This property indicates cumulative amounts of electric energy in 0.001 kWh.
	CumulativeElectricEnergyOfGeneration uint32
}

func decodeInstantaneousElectricPowerGeneration(edt []byte) (uint16, error) {
	if len(edt) != 2 {
		return 0, fmt.Errorf("expected 2 bytes, got %d", len(edt))
	}
	value := binary.BigEndian.Uint16(edt)
	if value > 65_533 {
		return 0, fmt.Errorf("electric power out of range: %d", value)
	}
	return value, nil
}

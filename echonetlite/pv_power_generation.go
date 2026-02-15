package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
)

const (
	epcInstantaneousElectricPowerGeneration = 0xE0
	epcCumulativeElectricEnergyOfGeneration = 0xE1
)

type PVPowerGeneration struct {
	host string
	eoj  EOJ
	conn *Connection
}

func NewPVPowerGeneration(host string, eoj EOJ, conn *Connection) *PVPowerGeneration {
	return &PVPowerGeneration{
		host: host,
		eoj:  eoj,
		conn: conn,
	}
}

func (p *PVPowerGeneration) Host() string {
	return p.host
}

func (p *PVPowerGeneration) EOJ() EOJ {
	return p.eoj
}

func (p *PVPowerGeneration) Get(ctx context.Context) (*PVPowerGenerationProps, error) {
	getReq := Frame{
		SEOJ: EOJ{0x05, 0xFF, 0x01},
		DEOJ: p.eoj,
		ESV:  0x62,
		Properties: []Property{
			{EPC: epcInstantaneousElectricPowerGeneration, EDT: []byte{}},
			{EPC: epcCumulativeElectricEnergyOfGeneration, EDT: []byte{}},
		},
	}
	resFrame, err := p.conn.unicast(ctx, p.host, getReq)
	if err != nil {
		return nil, err
	}

	props := &PVPowerGenerationProps{}
	for _, prop := range resFrame.Properties {
		switch prop.EPC {
		case epcInstantaneousElectricPowerGeneration:
			val, err := parseInstantaneousElectricPowerGeneration(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid instantaneousElectricPowerGeneration data: %w", err)
			}
			props.InstantaneousElectricPowerGeneration = val
		case epcCumulativeElectricEnergyOfGeneration:
			val, err := parseCumulativeElectricEnergyOfGeneration(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid cumulativeElectricEnergyOfGeneration data: %w", err)
			}
			props.CumulativeElectricEnergyOfGeneration = val
		}
	}

	return props, nil
}

type PVPowerGenerationProps struct {
	// This property indicates instantaneous generated power in watts.
	InstantaneousElectricPowerGeneration uint16

	// This property indicates cumulative amounts of electric energy in 0.001 kWh.
	CumulativeElectricEnergyOfGeneration uint32
}

func parseInstantaneousElectricPowerGeneration(edts []byte) (uint16, error) {
	if len(edts) != 2 {
		return 0, fmt.Errorf("expected 2 bytes, got %d", len(edts))
	}
	val := binary.BigEndian.Uint16(edts)
	return val, nil
}

func parseCumulativeElectricEnergyOfGeneration(edts []byte) (uint32, error) {
	if len(edts) != 4 {
		return 0, fmt.Errorf("expected 4 bytes, got %d", len(edts))
	}
	val := binary.BigEndian.Uint32(edts)
	return val, nil
}

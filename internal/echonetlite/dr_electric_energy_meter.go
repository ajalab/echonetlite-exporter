package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
)

const (
	epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit                 = 0xD4
	epcDRElectricEnergyMeterACInputCumulativeElectricEnergy                       = 0xE0
	epcDRElectricEnergyMeterACOutputCumulativeElectricEnergy                      = 0xE2
	epcDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy          = 0xE4
	epcDRElectricEnergyMeterACInstantaneousElectricPower                          = 0xE9
	epcDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower        = 0xEA
	drElectricEnergyMeterNoDataCumulative                                  uint32 = 0xFFFFFFFE
	drElectricEnergyMeterNoDataInstantaneous                               int32  = 0x7FFFFFFE
)

type DRElectricEnergyMeterClient struct {
	conn *Connection
}

func NewDRElectricEnergyMeterClient(conn *Connection) *DRElectricEnergyMeterClient {
	return &DRElectricEnergyMeterClient{conn: conn}
}

func (c *DRElectricEnergyMeterClient) Get(ctx context.Context, host string, eoj EOJ) (*DRElectricEnergyMeter, error) {
	getReq := Frame{
		SEOJ: EOJ{0x05, 0xFF, 0x01},
		DEOJ: eoj,
		ESV:  0x62,
		Properties: []Property{
			{EPC: epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit, EDT: []byte{}},
			{EPC: epcDRElectricEnergyMeterACInputCumulativeElectricEnergy, EDT: []byte{}},
			{EPC: epcDRElectricEnergyMeterACOutputCumulativeElectricEnergy, EDT: []byte{}},
			{EPC: epcDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy, EDT: []byte{}},
			{EPC: epcDRElectricEnergyMeterACInstantaneousElectricPower, EDT: []byte{}},
			{EPC: epcDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower, EDT: []byte{}},
		},
	}
	resFrame, err := c.conn.Unicast(ctx, host, getReq)
	if err != nil {
		return nil, err
	}
	if len(resFrame.Properties) != len(getReq.Properties) {
		return nil, fmt.Errorf("unexpected property count: got %d, want %d", len(resFrame.Properties), len(getReq.Properties))
	}

	meter := &DRElectricEnergyMeter{}
	for _, prop := range resFrame.Properties {
		switch prop.EPC {
		case epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit:
			meter.CumulativeAmountsOfElectricEnergyUnit = parseDRElectricEnergyMeterUnit(prop.EDT)
		case epcDRElectricEnergyMeterACInputCumulativeElectricEnergy:
			val, err := parseDRElectricEnergyMeterCumulative(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid acInputCumulativeElectricEnergy data: %w", err)
			}
			meter.ACInputCumulativeElectricEnergy = val
		case epcDRElectricEnergyMeterACOutputCumulativeElectricEnergy:
			val, err := parseDRElectricEnergyMeterCumulative(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid acOutputCumulativeElectricEnergy data: %w", err)
			}
			meter.ACOutputCumulativeElectricEnergy = val
		case epcDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy:
			val, err := parseDRElectricEnergyMeterCumulativeWithNoData(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid independentOperationCumulativeElectricEnergy data: %w", err)
			}
			meter.IndependentOperationCumulativeElectricEnergy = val
		case epcDRElectricEnergyMeterACInstantaneousElectricPower:
			val, err := parseDRElectricEnergyMeterInstantaneousPowerWithNoData(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid acInstantaneousElectricPower data: %w", err)
			}
			meter.ACInstantaneousElectricPower = val
		case epcDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower:
			val, err := parseDRElectricEnergyMeterInstantaneousPowerWithNoData(prop.EDT)
			if err != nil {
				return nil, fmt.Errorf("invalid independentOperationInstantaneousElectricPower data: %w", err)
			}
			meter.IndependentOperationInstantaneousElectricPower = val
		default:
			return nil, fmt.Errorf("unexpected EPC: 0x%X", prop.EPC)
		}
	}

	return meter, nil
}

type DRElectricEnergyMeter struct {
	CumulativeAmountsOfElectricEnergyUnit float32

	ACInputCumulativeElectricEnergy uint32

	ACOutputCumulativeElectricEnergy uint32

	IndependentOperationCumulativeElectricEnergy uint32

	ACInstantaneousElectricPower int32

	IndependentOperationInstantaneousElectricPower int32
}

func parseDRElectricEnergyMeterUnit(edts []byte) float32 {
	if len(edts) != 1 {
		return 1.0
	}
	switch edts[0] {
	case 0x00:
		return 1.0
	case 0x01:
		return 0.1
	case 0x02:
		return 0.01
	case 0x03:
		return 0.001
	case 0x04:
		return 0.0001
	case 0x0A:
		return 10.0
	case 0x0B:
		return 100.0
	case 0x0C:
		return 1000.0
	case 0x0D:
		return 10000.0
	default:
		return 1.0
	}
}

func parseDRElectricEnergyMeterCumulative(edts []byte) (uint32, error) {
	if len(edts) != 4 {
		return 0, fmt.Errorf("expected 4 bytes, got %d", len(edts))
	}
	return binary.BigEndian.Uint32(edts), nil
}

func parseDRElectricEnergyMeterCumulativeWithNoData(edts []byte) (uint32, error) {
	val, err := parseDRElectricEnergyMeterCumulative(edts)
	if err != nil {
		return 0, err
	}
	if val == drElectricEnergyMeterNoDataCumulative {
		return 0, fmt.Errorf("no data")
	}
	return val, nil
}

func parseDRElectricEnergyMeterInstantaneousPowerWithNoData(edts []byte) (int32, error) {
	if len(edts) != 4 {
		return 0, fmt.Errorf("expected 4 bytes, got %d", len(edts))
	}
	val := int32(binary.BigEndian.Uint32(edts))
	if val == drElectricEnergyMeterNoDataInstantaneous {
		return 0, fmt.Errorf("no data")
	}
	return val, nil
}

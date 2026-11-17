package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
)

const (
	epcWaterHeatingStatus                 = 0xB2
	epcHeaterStatus                       = 0xB3
	epcLinkageModeForSolarPowerGeneration = 0xB8
	epcHotWaterSupplyStatus               = 0xC3
	epcRemainingWater                     = 0xE1
	epcTankCapacity                       = 0xE2
)

type HybridWaterHeater struct {
	LinkageModeForSolarPowerGeneration uint8
	WaterHeatingStatus                 *bool
	HeaterStatus                       *bool
	HotWaterSupplyStatus               *bool
	RemainingWater                     *uint16 // L
	TankCapacity                       *uint16 // L
}

type HybridWaterHeaterClient struct {
	conn *Connection
}

func NewHybridWaterHeaterClient(conn *Connection) *HybridWaterHeaterClient {
	return &HybridWaterHeaterClient{conn: conn}
}

func (c *HybridWaterHeaterClient) Get(ctx context.Context, host string, eoj EOJ) (*HybridWaterHeater, error) {
	supported, err := getReadableProperties(ctx, c.conn, host, eoj)
	if err != nil {
		return nil, err
	}

	// AIF §3.1.4 allows these device properties to be read together.
	var optionalGroup []uint8
	for _, epc := range []uint8{epcWaterHeatingStatus, epcHotWaterSupplyStatus} {
		if supported[epc] {
			optionalGroup = append(optionalGroup, epc)
		}
	}
	values, err := getPropertyGroup(ctx, c.conn, host, eoj,
		[]uint8{epcLinkageModeForSolarPowerGeneration}, optionalGroup...)
	if err != nil {
		return nil, err
	}
	linkageMode := decodeByte(values[epcLinkageModeForSolarPowerGeneration], 0x41, 0x44)
	if linkageMode == nil {
		return nil, fmt.Errorf("invalid linkageModeForSolarPowerGeneration: %X", values[epcLinkageModeForSolarPowerGeneration])
	}
	optional := getOptionalPropertiesIndividually(ctx, c.conn, host, eoj, supported,
		epcHeaterStatus, epcRemainingWater, epcTankCapacity)
	return &HybridWaterHeater{
		LinkageModeForSolarPowerGeneration: *linkageMode,
		WaterHeatingStatus:                 decodeHeatingState(values[epcWaterHeatingStatus]),
		HeaterStatus:                       decodeHeatingState(optional[epcHeaterStatus]),
		HotWaterSupplyStatus:               decodeHeatingState(values[epcHotWaterSupplyStatus]),
		RemainingWater:                     decodeWaterVolume(optional[epcRemainingWater]),
		TankCapacity:                       decodeWaterVolume(optional[epcTankCapacity]),
	}, nil
}

func decodeWaterVolume(edt []byte) *uint16 {
	if len(edt) != 2 {
		return nil
	}
	value := binary.BigEndian.Uint16(edt)
	if value > 0xFFFD {
		return nil
	}
	return &value
}

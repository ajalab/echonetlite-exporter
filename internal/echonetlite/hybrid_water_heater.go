package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
	"maps"
	"time"
)

const (
	EPCWaterHeatingStatus                 = 0xB2
	EPCHeaterStatus                       = 0xB3
	EPCLinkageModeForSolarPowerGeneration = 0xB8
	EPCHotWaterSupplyStatus               = 0xC3
	EPCRemainingWater                     = 0xE1
	EPCTankCapacity                       = 0xE2
)

type HybridWaterHeater struct {
	SuperClass

	LinkageModeForSolarPowerGeneration uint8
	WaterHeatingStatus                 *uint8
	HeaterStatus                       *uint8
	HotWaterSupplyStatus               *uint8
	RemainingWater                     *uint16 // L
	TankCapacity                       *uint16 // L

	AutomaticWaterHeating                           *uint8
	HotWaterSupplyModeForAuxiliaryHeatSourceMachine *uint8
	HeaterModeForAuxiliaryHeatSourceMachine         *uint8
}

type HybridWaterHeaterClient struct {
	conn *Connection
}

func NewHybridWaterHeaterClient(conn *Connection) *HybridWaterHeaterClient {
	return &HybridWaterHeaterClient{conn: conn}
}

func (c *HybridWaterHeaterClient) Get(ctx context.Context, host string, eoj EOJ) (*HybridWaterHeater, error) {
	read := propertyReader{conn: c.conn, host: host, eoj: eoj, minimumWait: 10 * time.Second}
	supported, err := read.readableProperties(ctx)
	if err != nil {
		return nil, err
	}

	// AIF §3.1.4 allows these device properties to be read together.
	values, err := read.group(ctx, []uint8{EPCLinkageModeForSolarPowerGeneration},
		supportedEPCs(supported, EPCWaterHeatingStatus, EPCHotWaterSupplyStatus))
	if err != nil {
		return nil, err
	}
	linkageMode, err := decodeByte(values[EPCLinkageModeForSolarPowerGeneration], 0x41, 0x44)
	if err != nil {
		return nil, fmt.Errorf("linkageModeForSolarPowerGeneration (EPC 0xB8): %w", err)
	}
	optional, err := read.optionalIndividually(ctx, supportedEPCs(supported,
		EPCHeaterStatus, EPCRemainingWater, EPCTankCapacity))
	if err != nil {
		return nil, err
	}
	maps.Copy(values, optional)
	reading := &HybridWaterHeater{LinkageModeForSolarPowerGeneration: linkageMode}
	for _, field := range []struct {
		epc  uint8
		dest **uint8
	}{
		{EPCWaterHeatingStatus, &reading.WaterHeatingStatus},
		{EPCHeaterStatus, &reading.HeaterStatus},
		{EPCHotWaterSupplyStatus, &reading.HotWaterSupplyStatus},
	} {
		edt := values[field.epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeByte(edt, 0x41, 0x42)
		if err != nil {
			return nil, fmt.Errorf("heating state (EPC 0x%02X): %w", field.epc, err)
		}
		*field.dest = &value
	}
	for _, field := range []struct {
		epc  uint8
		dest **uint16
	}{
		{EPCRemainingWater, &reading.RemainingWater},
		{EPCTankCapacity, &reading.TankCapacity},
	} {
		edt := values[field.epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeWaterVolume(edt)
		if err != nil {
			return nil, fmt.Errorf("water volume (EPC 0x%02X): %w", field.epc, err)
		}
		*field.dest = &value
	}

	monitoringRequired, monitoringOptional := superClassEPCs(supported)
	monitoringValues, err := read.group(ctx, monitoringRequired, monitoringOptional)
	if err != nil {
		return nil, err
	}
	individual := supportedEPCs(supported, 0xB0, 0xB6, 0xB7)
	// Retry optional EPCs omitted from a short Get_SNA.
	for _, epc := range monitoringOptional {
		if _, present := monitoringValues[epc]; !present {
			individual = append(individual, epc)
		}
	}
	if len(individual) != 0 {
		additional, err := read.optionalIndividually(ctx, individual)
		if err != nil {
			return nil, err
		}
		maps.Copy(monitoringValues, additional)
	}

	reading.SuperClass, err = decodeSuperClass(monitoringValues)
	if err != nil {
		return nil, err
	}

	for _, field := range []struct {
		epc      uint8
		min, max uint8
		dest     **uint8
	}{
		{0xB0, 0x41, 0x43, &reading.AutomaticWaterHeating},
		{0xB6, 0x41, 0x42, &reading.HotWaterSupplyModeForAuxiliaryHeatSourceMachine},
		{0xB7, 0x41, 0x42, &reading.HeaterModeForAuxiliaryHeatSourceMachine},
	} {
		edt := monitoringValues[field.epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeByte(edt, field.min, field.max)
		if err != nil {
			return nil, fmt.Errorf("EPC 0x%02X: %w", field.epc, err)
		}
		*field.dest = &value
	}
	return reading, nil
}

func decodeWaterVolume(edt []byte) (uint16, error) {
	if len(edt) != 2 {
		return 0, fmt.Errorf("expected 2 bytes, got %d", len(edt))
	}
	value := binary.BigEndian.Uint16(edt)
	if value > 0xFFFD {
		return 0, fmt.Errorf("water volume out of range: %d", value)
	}
	return value, nil
}

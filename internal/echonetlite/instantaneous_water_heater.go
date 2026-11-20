package echonetlite

import (
	"context"
	"fmt"
	"maps"
	"time"
)

const (
	epcHotWaterHeatingStatus          = 0xD0
	epcTargetSuppliedWaterTemperature = 0xD1
	epcTargetBathWaterTemperature     = 0xE1
	epcBathWaterHeatingStatus         = 0xE2
	epcAutomaticBathOperation         = 0xE3
	epcShowerHotWaterSupplyStatus     = 0xEA
	epcKitchenHotWaterSupplyStatus    = 0xEB
	epcBathOperationStatus            = 0xEF
)

type InstantaneousWaterHeater struct {
	HotWaterHeatingStatus          bool
	BathWaterHeatingStatus         bool
	AutomaticBathOperation         *bool
	TargetSuppliedWaterTemperature *uint8 // °C
	TargetBathWaterTemperature     *uint8 // °C
	ShowerHotWaterSupplyStatus     *bool
	KitchenHotWaterSupplyStatus    *bool
	BathOperationStatus            *uint8
}

type InstantaneousWaterHeaterClient struct {
	conn *Connection
}

func NewInstantaneousWaterHeaterClient(conn *Connection) *InstantaneousWaterHeaterClient {
	return &InstantaneousWaterHeaterClient{conn: conn}
}

func (c *InstantaneousWaterHeaterClient) Get(ctx context.Context, host string, eoj EOJ) (*InstantaneousWaterHeater, error) {
	read := propertyReader{conn: c.conn, host: host, eoj: eoj, minimumWait: 10 * time.Second}
	supported, err := read.readableProperties(ctx)
	if err != nil {
		return nil, err
	}

	// AIF §3.1.4 allows these device properties to be read together.
	values, err := read.group(ctx, []uint8{epcHotWaterHeatingStatus, epcBathWaterHeatingStatus},
		supportedEPCs(supported, epcAutomaticBathOperation))
	if err != nil {
		return nil, err
	}
	hotWaterHeating, err := decodeHeatingState(values[epcHotWaterHeatingStatus])
	if err != nil {
		return nil, fmt.Errorf("hotWaterHeatingStatus (EPC 0xD0): %w", err)
	}
	bathWaterHeating, err := decodeHeatingState(values[epcBathWaterHeatingStatus])
	if err != nil {
		return nil, fmt.Errorf("bathWaterHeatingStatus (EPC 0xE2): %w", err)
	}

	optional, err := read.optionalIndividually(ctx, supportedEPCs(supported,
		epcTargetSuppliedWaterTemperature, epcTargetBathWaterTemperature,
		epcShowerHotWaterSupplyStatus, epcKitchenHotWaterSupplyStatus, epcBathOperationStatus))
	if err != nil {
		return nil, err
	}
	maps.Copy(values, optional)
	reading := &InstantaneousWaterHeater{
		HotWaterHeatingStatus:  hotWaterHeating,
		BathWaterHeatingStatus: bathWaterHeating,
	}
	for _, field := range []struct {
		epc  uint8
		dest **bool
	}{
		{epcAutomaticBathOperation, &reading.AutomaticBathOperation},
		{epcShowerHotWaterSupplyStatus, &reading.ShowerHotWaterSupplyStatus},
		{epcKitchenHotWaterSupplyStatus, &reading.KitchenHotWaterSupplyStatus},
	} {
		edt := values[field.epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeHeatingState(edt)
		if err != nil {
			return nil, fmt.Errorf("heating state (EPC 0x%02X): %w", field.epc, err)
		}
		*field.dest = &value
	}
	for _, field := range []struct {
		epc      uint8
		min, max uint8
		dest     **uint8
	}{
		{epcTargetSuppliedWaterTemperature, 0, 100, &reading.TargetSuppliedWaterTemperature},
		{epcTargetBathWaterTemperature, 0, 100, &reading.TargetBathWaterTemperature},
		{epcBathOperationStatus, 0x41, 0x43, &reading.BathOperationStatus},
	} {
		edt := values[field.epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeByte(edt, field.min, field.max)
		if err != nil {
			return nil, fmt.Errorf("water heater property (EPC 0x%02X): %w", field.epc, err)
		}
		*field.dest = &value
	}
	return reading, nil
}

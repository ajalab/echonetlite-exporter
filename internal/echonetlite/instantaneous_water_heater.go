package echonetlite

import (
	"context"
	"fmt"
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
	supported, err := getReadableProperties(ctx, c.conn, host, eoj)
	if err != nil {
		return nil, err
	}

	// AIF §3.1.4 allows these device properties to be read together.
	var optionalGroup []uint8
	if supported[epcAutomaticBathOperation] {
		optionalGroup = append(optionalGroup, epcAutomaticBathOperation)
	}
	values, err := getPropertyGroup(ctx, c.conn, host, eoj,
		[]uint8{epcHotWaterHeatingStatus, epcBathWaterHeatingStatus}, optionalGroup...)
	if err != nil {
		return nil, err
	}
	hotWaterHeating := decodeHeatingState(values[epcHotWaterHeatingStatus])
	if hotWaterHeating == nil {
		return nil, fmt.Errorf("invalid hotWaterHeatingStatus: %X", values[epcHotWaterHeatingStatus])
	}
	bathWaterHeating := decodeHeatingState(values[epcBathWaterHeatingStatus])
	if bathWaterHeating == nil {
		return nil, fmt.Errorf("invalid bathWaterHeatingStatus: %X", values[epcBathWaterHeatingStatus])
	}

	optional := getOptionalPropertiesIndividually(ctx, c.conn, host, eoj, supported,
		epcTargetSuppliedWaterTemperature, epcTargetBathWaterTemperature,
		epcShowerHotWaterSupplyStatus, epcKitchenHotWaterSupplyStatus, epcBathOperationStatus)
	return &InstantaneousWaterHeater{
		HotWaterHeatingStatus:          *hotWaterHeating,
		BathWaterHeatingStatus:         *bathWaterHeating,
		AutomaticBathOperation:         decodeHeatingState(values[epcAutomaticBathOperation]),
		TargetSuppliedWaterTemperature: decodeByte(optional[epcTargetSuppliedWaterTemperature], 0, 100),
		TargetBathWaterTemperature:     decodeByte(optional[epcTargetBathWaterTemperature], 0, 100),
		ShowerHotWaterSupplyStatus:     decodeHeatingState(optional[epcShowerHotWaterSupplyStatus]),
		KitchenHotWaterSupplyStatus:    decodeHeatingState(optional[epcKitchenHotWaterSupplyStatus]),
		BathOperationStatus:            decodeByte(optional[epcBathOperationStatus], 0x41, 0x43),
	}, nil
}

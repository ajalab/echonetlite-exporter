package echonetlite

import (
	"context"
	"fmt"
	"maps"
	"time"
)

const (
	epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit          = 0xD4
	EPCDRElectricEnergyMeterACInputCumulativeElectricEnergy                = 0xE0
	EPCDRElectricEnergyMeterACOutputCumulativeElectricEnergy               = 0xE2
	EPCDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy   = 0xE4
	EPCDRElectricEnergyMeterACInstantaneousElectricPower                   = 0xE9
	EPCDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower = 0xEA
)

type DRElectricEnergyMeterClient struct {
	conn *Connection
}

func NewDRElectricEnergyMeterClient(conn *Connection) *DRElectricEnergyMeterClient {
	return &DRElectricEnergyMeterClient{conn: conn}
}

type DRElectricEnergyMeter struct {
	SuperClass

	CumulativeAmountsOfElectricEnergyUnit float64 // kWh per count

	// At least one AC energy direction is supported. Every supported direction is required.
	ACInputCumulativeElectricEnergy  *uint32
	ACOutputCumulativeElectricEnergy *uint32

	// These are optional. A valid NoData code also leaves the property absent.
	IndependentOperationCumulativeElectricEnergy   *uint32
	ACInstantaneousElectricPower                   *int32
	IndependentOperationInstantaneousElectricPower *int32

	TimeSynchronizationStatus *uint8
}

func (c *DRElectricEnergyMeterClient) Get(ctx context.Context, host string, eoj EOJ) (*DRElectricEnergyMeter, error) {
	read := propertyReader{conn: c.conn, host: host, eoj: eoj, minimumWait: 20 * time.Second}
	supported, err := read.readableProperties(ctx)
	if err != nil {
		return nil, err
	}
	required := supportedEPCs(supported, EPCDRElectricEnergyMeterACInputCumulativeElectricEnergy, EPCDRElectricEnergyMeterACOutputCumulativeElectricEnergy)
	if len(required) == 0 {
		return nil, fmt.Errorf("Get property map lacks both AC energy directions (EPC 0xE0 and 0xE2)")
	}

	// AIF §3.1.4 lists D4 as an attribute. §3.2.1 groups the current values.
	unitValues, err := read.required(ctx, epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit)
	if err != nil {
		return nil, err
	}
	unit, err := decodeElectricEnergyUnit(unitValues[epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit])
	if err != nil {
		return nil, fmt.Errorf("cumulativeAmountsOfElectricEnergyUnit (EPC 0xD4): %w", err)
	}
	values, err := read.group(ctx, required, supportedEPCs(supported,
		EPCDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy, EPCDRElectricEnergyMeterACInstantaneousElectricPower, EPCDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower))
	if err != nil {
		return nil, err
	}

	reading := &DRElectricEnergyMeter{CumulativeAmountsOfElectricEnergyUnit: unit}
	for _, field := range []struct {
		epc  uint8
		dest **uint32
	}{
		{EPCDRElectricEnergyMeterACInputCumulativeElectricEnergy, &reading.ACInputCumulativeElectricEnergy},
		{EPCDRElectricEnergyMeterACOutputCumulativeElectricEnergy, &reading.ACOutputCumulativeElectricEnergy},
		{EPCDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy, &reading.IndependentOperationCumulativeElectricEnergy},
	} {
		edt := values[field.epc]
		if len(edt) == 0 || (field.epc == EPCDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy && isNoData(edt, noEnergyData)) {
			continue
		}
		value, err := decodeMeterValue(edt, 999_999_999)
		if err != nil {
			return nil, fmt.Errorf("DR meter energy (EPC 0x%02X): %w", field.epc, err)
		}
		*field.dest = &value
	}
	for _, field := range []struct {
		epc  uint8
		dest **int32
	}{
		{EPCDRElectricEnergyMeterACInstantaneousElectricPower, &reading.ACInstantaneousElectricPower},
		{EPCDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower, &reading.IndependentOperationInstantaneousElectricPower},
	} {
		edt := values[field.epc]
		if len(edt) == 0 || isNoData(edt, noPowerData) {
			continue
		}
		value, err := decodeElectricPower(edt)
		if err != nil {
			return nil, fmt.Errorf("DR meter power (EPC 0x%02X): %w", field.epc, err)
		}
		*field.dest = &value
	}

	monitoringRequired, monitoringOptional := superClassEPCs(supported, 0xDB)
	monitoringValues, err := read.group(ctx, monitoringRequired, monitoringOptional)
	if err != nil {
		return nil, err
	}
	var individual []uint8
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

	if edt := monitoringValues[0xDB]; len(edt) != 0 {
		if len(edt) != 1 {
			return nil, fmt.Errorf("EPC 0xDB: expected 1 byte, got %d", len(edt))
		}
		switch edt[0] {
		case 0x00, 0x01, 0x02, 0xFF:
			reading.TimeSynchronizationStatus = &edt[0]
		default:
			return nil, fmt.Errorf("EPC 0xDB: invalid state code 0x%02X", edt[0])
		}
	}
	return reading, nil
}

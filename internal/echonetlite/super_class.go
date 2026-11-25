package echonetlite

import (
	"encoding/binary"
	"fmt"
	"slices"
)

const (
	EPCOperationStatus                       = 0x80
	EPCInstantaneousElectricPowerConsumption = 0x84
	EPCConsumedCumulativeElectricEnergy      = 0x85
	EPCCurrentLimit                          = 0x87
	EPCFaultStatus                           = 0x88
	EPCFaultDescription                      = 0x89
	EPCPowerSaving                           = 0x8F
	EPCPowerLimit                            = 0x99
	EPCHourMeter                             = 0x9A
)

// SuperClass contains the monitored properties defined by the device object
// superclass. A nil field means the reading has no sample for that property.
type SuperClass struct {
	OperationStatus                       *uint8
	InstantaneousElectricPowerConsumption *uint16 // W
	ConsumedCumulativeElectricEnergy      *uint32 // 0.001 kWh
	CurrentLimit                          *uint8  // %
	FaultStatus                           *uint8
	FaultDescription                      *uint16
	PowerSaving                           *uint8
	PowerLimit                            *uint16 // W
	HourMeter                             *uint64 // seconds
}

// superClassEPCs combines the superclass defaults with EPCs the class requires.
// A class may make a superclass property mandatory, such as fault description.
func superClassEPCs(supported [256]bool, classRequired ...uint8) (required, optional []uint8) {
	required = append([]uint8{EPCOperationStatus, EPCFaultStatus}, classRequired...)
	for _, epc := range []uint8{
		EPCInstantaneousElectricPowerConsumption, EPCConsumedCumulativeElectricEnergy, EPCCurrentLimit,
		EPCFaultDescription, EPCPowerSaving, EPCPowerLimit, EPCHourMeter,
	} {
		if supported[epc] && !slices.Contains(required, epc) {
			optional = append(optional, epc)
		}
	}
	return required, optional
}

func decodeSuperClass(values map[uint8][]byte) (SuperClass, error) {
	operationStatus, err := decodeByte(values[EPCOperationStatus], 0x30, 0x31)
	if err != nil {
		return SuperClass{}, fmt.Errorf("EPC 0x%02X: %w", EPCOperationStatus, err)
	}
	faultStatus, err := decodeByte(values[EPCFaultStatus], 0x41, 0x42)
	if err != nil {
		return SuperClass{}, fmt.Errorf("EPC 0x%02X: %w", EPCFaultStatus, err)
	}
	reading := SuperClass{OperationStatus: &operationStatus, FaultStatus: &faultStatus}
	for _, field := range []struct {
		epc      uint8
		min, max uint8
		dest     **uint8
	}{
		{EPCCurrentLimit, 0, 100, &reading.CurrentLimit},
		{EPCPowerSaving, 0x41, 0x42, &reading.PowerSaving},
	} {
		edt := values[field.epc]
		if len(edt) == 0 {
			continue
		}
		value, err := decodeByte(edt, field.min, field.max)
		if err != nil {
			return SuperClass{}, fmt.Errorf("EPC 0x%02X: %w", field.epc, err)
		}
		*field.dest = &value
	}
	for _, field := range []struct {
		epc  uint8
		dest **uint16
	}{
		{EPCInstantaneousElectricPowerConsumption, &reading.InstantaneousElectricPowerConsumption},
		{EPCFaultDescription, &reading.FaultDescription},
		{EPCPowerLimit, &reading.PowerLimit},
	} {
		edt := values[field.epc]
		if len(edt) == 0 {
			continue
		}
		if len(edt) != 2 {
			return SuperClass{}, fmt.Errorf("EPC 0x%02X: expected 2 bytes, got %d", field.epc, len(edt))
		}
		value := binary.BigEndian.Uint16(edt)
		// Appendix §2.20 reserves these codes for underflow and overflow.
		// The power limit setting uses the full uint16 range (§2.22).
		if field.epc == EPCInstantaneousElectricPowerConsumption && value >= 0xFFFE {
			continue
		}
		*field.dest = &value
	}
	if edt := values[EPCConsumedCumulativeElectricEnergy]; len(edt) != 0 {
		value, err := decodeMeterValue(edt, 999_999_999)
		if err != nil {
			return SuperClass{}, fmt.Errorf("EPC 0x%02X: %w", EPCConsumedCumulativeElectricEnergy, err)
		}
		reading.ConsumedCumulativeElectricEnergy = &value
	}
	if edt := values[EPCHourMeter]; len(edt) != 0 {
		if len(edt) != 5 {
			return SuperClass{}, fmt.Errorf("EPC 0x%02X: expected 5 bytes, got %d", EPCHourMeter, len(edt))
		}
		unit, err := decodeByte(edt[:1], 0x41, 0x44)
		if err != nil {
			return SuperClass{}, fmt.Errorf("EPC 0x%02X: %w", EPCHourMeter, err)
		}
		count := binary.BigEndian.Uint32(edt[1:])
		if count != 0xFFFFFFFF { // Appendix §2.17 defines this as overflow.
			secondsPerUnit := [...]uint64{1, 60, 3600, 86400}
			reading.HourMeter = new(uint64(count) * secondsPerUnit[unit-0x41])
		}
	}
	return reading, nil
}

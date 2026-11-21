package echonetlite

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	noEnergyData = 0xFFFFFFFE
	noPowerData  = 0x7FFFFFFE
)

func decodeMeterValue(edt []byte, max uint32) (uint32, error) {
	if len(edt) != 4 {
		return 0, fmt.Errorf("expected 4 bytes, got %d", len(edt))
	}
	value := binary.BigEndian.Uint32(edt)
	if value > max {
		return 0, fmt.Errorf("value %d exceeds %d", value, max)
	}
	return value, nil
}

func decodeByte(edt []byte, min, max uint8) (uint8, error) {
	if len(edt) != 1 {
		return 0, fmt.Errorf("expected 1 byte, got %d", len(edt))
	}
	if edt[0] < min || edt[0] > max {
		return 0, fmt.Errorf("value 0x%02X outside 0x%02X–0x%02X", edt[0], min, max)
	}
	return edt[0], nil
}

func decodeHeatingState(edt []byte) (bool, error) {
	value, err := decodeByte(edt, 0x41, 0x42)
	if err != nil {
		return false, err
	}
	return value == 0x41, nil
}

func decodeElectricEnergyUnit(edt []byte) (float64, error) {
	if len(edt) == 1 {
		switch code := edt[0]; {
		case code <= 0x04:
			return math.Pow10(-int(code)), nil
		case code >= 0x0A && code <= 0x0D:
			return math.Pow10(int(code) - 9), nil
		}
	}
	return 0, fmt.Errorf("invalid electric energy unit: %X", edt)
}

func decodeElectricPower(edt []byte) (int32, error) {
	if len(edt) != 4 {
		return 0, fmt.Errorf("expected 4 bytes, got %d", len(edt))
	}
	power := int32(binary.BigEndian.Uint32(edt))
	if power < -2_147_483_647 || power > 2_147_483_645 {
		return 0, fmt.Errorf("electric power out of range: %d", power)
	}
	return power, nil
}

func isNoData(edt []byte, code uint32) bool {
	return len(edt) == 4 && binary.BigEndian.Uint32(edt) == code
}

package echonetlite

import (
	"context"
	"fmt"
)

const epcGetPropertyMap = 0x9F

func (r propertyReader) readableProperties(ctx context.Context) ([256]bool, error) {
	values, err := r.required(ctx, epcGetPropertyMap)
	if err != nil {
		return [256]bool{}, err
	}
	supported, err := decodePropertyMap(values[epcGetPropertyMap])
	if err != nil {
		return [256]bool{}, fmt.Errorf("invalid Get property map: %w", err)
	}
	return supported, nil
}

func decodePropertyMap(edt []byte) ([256]bool, error) {
	var supported [256]bool
	if len(edt) == 0 {
		return supported, fmt.Errorf("missing property count")
	}
	count := int(edt[0])
	if count < 16 {
		if len(edt) != count+1 {
			return supported, fmt.Errorf("expected %d EPCs, got %d", count, len(edt)-1)
		}
		for _, epc := range edt[1:] {
			if epc < 0x80 || supported[epc] {
				return supported, fmt.Errorf("invalid EPC 0x%02X in property list", epc)
			}
			supported[epc] = true
		}
		return supported, nil
	}
	if len(edt) != 17 {
		return supported, fmt.Errorf("expected a 16-byte bitmap, got %d bytes", len(edt)-1)
	}
	var actualCount int
	for lowNibble, bitmap := range edt[1:] {
		for bit := range 8 {
			if bitmap&(1<<bit) != 0 {
				supported[(bit+8)<<4|lowNibble] = true
				actualCount++
			}
		}
	}
	if actualCount != count {
		return supported, fmt.Errorf("expected %d EPCs in bitmap, got %d", count, actualCount)
	}
	return supported, nil
}

// supportedEPCs preserves the request order supplied by the client.
func supportedEPCs(supported [256]bool, candidates ...uint8) []uint8 {
	var epcs []uint8
	for _, epc := range candidates {
		if supported[epc] {
			epcs = append(epcs, epc)
		}
	}
	return epcs
}

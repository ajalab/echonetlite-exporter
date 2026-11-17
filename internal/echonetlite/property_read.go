package echonetlite

import (
	"context"
	"encoding/binary"
	"fmt"
	"slices"
)

const epcGetPropertyMap = 0x9F

func getRequiredProperties(ctx context.Context, conn *Connection, host string, eoj EOJ, epcs ...uint8) (map[uint8][]byte, error) {
	return getPropertyGroup(ctx, conn, host, eoj, epcs)
}

func getPropertyGroup(ctx context.Context, conn *Connection, host string, eoj EOJ, required []uint8, optional ...uint8) (map[uint8][]byte, error) {
	values, err := requestProperties(ctx, conn, host, eoj, slices.Concat(required, optional))
	if err != nil {
		return nil, err
	}
	for _, epc := range required {
		if len(values[epc]) == 0 {
			return nil, fmt.Errorf("missing required EPC 0x%02X", epc)
		}
	}
	return values, nil
}

func requestProperties(ctx context.Context, conn *Connection, host string, eoj EOJ, epcs []uint8) (map[uint8][]byte, error) {
	req := Frame{SEOJ: EOJ{0x05, 0xFF, 0x01}, DEOJ: eoj, ESV: 0x62}
	for _, epc := range epcs {
		req.Properties = append(req.Properties, Property{EPC: epc})
	}
	res, err := conn.Unicast(ctx, host, req)
	if err != nil {
		return nil, err
	}
	// Get_SNA can carry EDTs for properties that were read successfully.
	if res.SEOJ != req.DEOJ || res.DEOJ != req.SEOJ || (res.ESV != 0x72 && res.ESV != 0x52) {
		return nil, fmt.Errorf("unexpected Get response from %s to %s with ESV 0x%02X", res.SEOJ, res.DEOJ, res.ESV)
	}

	values := make(map[uint8][]byte, len(epcs))
	complete := len(res.Properties) == len(epcs)
	for _, prop := range res.Properties {
		if !slices.Contains(epcs, prop.EPC) {
			return nil, fmt.Errorf("unexpected EPC 0x%02X", prop.EPC)
		}
		if _, exists := values[prop.EPC]; exists {
			return nil, fmt.Errorf("duplicate EPC 0x%02X", prop.EPC)
		}
		values[prop.EPC] = prop.EDT
		complete = complete && len(prop.EDT) != 0
	}
	if res.ESV == 0x52 && complete {
		return nil, fmt.Errorf("Get_SNA response contains all requested values")
	}
	return values, nil
}

func getReadableProperties(ctx context.Context, conn *Connection, host string, eoj EOJ) ([256]bool, error) {
	values, err := getRequiredProperties(ctx, conn, host, eoj, epcGetPropertyMap)
	if err != nil {
		return [256]bool{}, err
	}
	supported, err := decodePropertyMap(values[epcGetPropertyMap])
	if err != nil {
		return [256]bool{}, fmt.Errorf("invalid Get property map: %w", err)
	}
	return supported, nil
}

func getOptionalPropertiesIndividually(ctx context.Context, conn *Connection, host string, eoj EOJ, supported [256]bool, epcs ...uint8) map[uint8][]byte {
	values := make(map[uint8][]byte, len(epcs))
	for _, epc := range epcs {
		if ctx.Err() != nil {
			break
		}
		if !supported[epc] {
			continue
		}
		reply, err := requestProperties(ctx, conn, host, eoj, []uint8{epc})
		if err == nil {
			values[epc] = reply[epc]
		}
	}
	return values
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

func decodeByte(edt []byte, min, max uint8) *uint8 {
	if len(edt) != 1 || edt[0] < min || edt[0] > max {
		return nil
	}
	value := edt[0]
	return &value
}

func decodeHeatingState(edt []byte) *bool {
	value := decodeByte(edt, 0x41, 0x42)
	if value == nil {
		return nil
	}
	on := *value == 0x41
	return &on
}

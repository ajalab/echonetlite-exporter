package echonetlite

import (
	"context"
	"fmt"
	"slices"
	"time"
)

const (
	esvGet    = 0x62
	esvGetRes = 0x72
	esvGetSNA = 0x52
)

var controllerEOJ = EOJ{0x05, 0xFF, 0x01}

type propertyReader struct {
	conn        *Connection
	host        string
	eoj         EOJ
	minimumWait time.Duration
}

func (r propertyReader) required(ctx context.Context, epcs ...uint8) (map[uint8][]byte, error) {
	return r.group(ctx, epcs, nil)
}

func (r propertyReader) group(ctx context.Context, required, optional []uint8) (map[uint8][]byte, error) {
	values, err := r.get(ctx, slices.Concat(required, optional))
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

// optionalIndividually reads the optional EPCs selected by the caller, one per request.
// Get_SNA without an EDT leaves that property absent.
func (r propertyReader) optionalIndividually(ctx context.Context, epcs []uint8) (map[uint8][]byte, error) {
	values := make(map[uint8][]byte, len(epcs))
	for _, epc := range epcs {
		reply, err := r.get(ctx, []uint8{epc})
		if err != nil {
			return nil, fmt.Errorf("read EPC 0x%02X: %w", epc, err)
		}
		values[epc] = reply[epc]
	}
	return values, nil
}

func (r propertyReader) get(ctx context.Context, epcs []uint8) (map[uint8][]byte, error) {
	req := newGetRequest(r.eoj, epcs)
	res, err := r.conn.unicast(ctx, r.host, req, r.minimumWait)
	if err != nil {
		return nil, err
	}
	return getResponseValues(req, res)
}

func newGetRequest(eoj EOJ, epcs []uint8) Frame {
	req := Frame{SEOJ: controllerEOJ, DEOJ: eoj, ESV: esvGet}
	for _, epc := range epcs {
		req.Properties = append(req.Properties, Property{EPC: epc})
	}
	return req
}

func getResponseValues(req Frame, res *Frame) (map[uint8][]byte, error) {
	if res.SEOJ != req.DEOJ || res.DEOJ != req.SEOJ || (res.ESV != esvGetRes && res.ESV != esvGetSNA) {
		return nil, fmt.Errorf("unexpected Get response from %s to %s with ESV 0x%02X", res.SEOJ, res.DEOJ, res.ESV)
	}

	values := make(map[uint8][]byte, len(res.Properties))
	shortSNA := res.ESV == esvGetSNA && len(res.Properties) < len(req.Properties)
	unavailable := shortSNA
	for i, prop := range res.Properties {
		if !slices.ContainsFunc(req.Properties, func(request Property) bool { return request.EPC == prop.EPC }) {
			return nil, fmt.Errorf("unexpected EPC 0x%02X", prop.EPC)
		}
		if _, exists := values[prop.EPC]; exists {
			return nil, fmt.Errorf("duplicate EPC 0x%02X", prop.EPC)
		}
		// §4.2.3.3 permits a prefix when a Get_SNA cannot fit or process the full response.
		if shortSNA && prop.EPC != req.Properties[i].EPC {
			return nil, fmt.Errorf("EPC 0x%02X out of order in short Get_SNA", prop.EPC)
		}
		if len(prop.EDT) == 0 {
			if res.ESV == esvGetRes {
				return nil, fmt.Errorf("empty Get_Res EDT for EPC 0x%02X", prop.EPC)
			}
			unavailable = true
		}
		values[prop.EPC] = prop.EDT
	}
	if res.ESV == esvGetRes && len(values) != len(req.Properties) {
		return nil, fmt.Errorf("unexpected property count: got %d, want %d", len(values), len(req.Properties))
	}
	if res.ESV == esvGetSNA && !unavailable {
		return nil, fmt.Errorf("Get_SNA response contains all requested values")
	}
	return values, nil
}

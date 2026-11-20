package echonetlite

import (
	"testing"
)

func TestGasMeterGet(t *testing.T) {
	eoj := EOJ{0x02, 0x82, 0x01}
	conn := newPropertyTestConnection(t, eoj, propertyExchange{
		[]uint8{0xE0}, 0x72, []Property{testProperty(0xE0, 0, 0, 0, 100)},
	})
	got, err := NewGasMeterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil || got.CumulativeGasConsumption != 100 {
		t.Fatalf("Get = (%+v, %v), want consumption count 100", got, err)
	}
}

func TestGasMeterGetRejectsInvalidConsumption(t *testing.T) {
	for _, tc := range []struct {
		name string
		esv  uint8
		edt  []byte
	}{
		{"unavailable", 0x52, nil},
		{"wrong length", 0x72, []byte{0, 0, 1}},
		{"above maximum", 0x72, []byte{0x3B, 0x9A, 0xCA, 0x00}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0x82, 0x01}
			conn := newPropertyTestConnection(t, eoj,
				propertyExchange{[]uint8{0xE0}, tc.esv, []Property{testProperty(0xE0, tc.edt...)}},
			)
			if got, err := NewGasMeterClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}
}

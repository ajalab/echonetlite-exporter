package echonetlite

import (
	"slices"
	"testing"
)

func TestMultipleInputPCSGet(t *testing.T) {
	eoj := EOJ{0x02, 0xA5, 0x01}
	epcs := []uint8{0xE0, 0xE3, 0xE7}
	response := []Property{
		testProperty(0xE0, 0, 0, 0, 200),
		testProperty(0xE3, 0, 0, 0, 100),
		testProperty(0xE7, 0xFF, 0xFF, 0xFF, 0x9C),
	}
	want := MultipleInputPCS{
		NormalDirectionElectricEnergy: 200, ReverseDirectionElectricEnergy: 100, InstantaneousElectricPower: -100,
	}

	t.Run("success", func(t *testing.T) {
		conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x72, response})
		got, err := NewMultipleInputPCSClient(conn).Get(t.Context(), "127.0.0.1", eoj)
		if err != nil || got == nil || *got != want {
			t.Fatalf("Get = (%+v, %v), want %+v", got, err, want)
		}
	})

	for _, tc := range []struct {
		name     string
		index    int
		property Property
	}{
		{"short reverse energy", 1, testProperty(0xE3, 0)},
		{"normal energy above max", 0, testProperty(0xE0, 0x3B, 0x9A, 0xCA, 0x00)},
		{"short power", 2, testProperty(0xE7, 0, 0, 0)},
		{"power underflow", 2, testProperty(0xE7, 0x80, 0, 0, 0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			props := slices.Clone(response)
			props[tc.index] = tc.property
			conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x72, props})
			if got, err := NewMultipleInputPCSClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}

	t.Run("missing required property", func(t *testing.T) {
		props := slices.Clone(response)
		props[len(props)-1].EDT = nil
		conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x52, props})
		if got, err := NewMultipleInputPCSClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
			t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
		}
	})
}

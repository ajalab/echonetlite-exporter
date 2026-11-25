package echonetlite

import (
	"reflect"
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
		SuperClass:                       SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		PowerSystemInterconnectionStatus: new(uint8(2)),
		NormalDirectionElectricEnergy:    200, ReverseDirectionElectricEnergy: 100, InstantaneousElectricPower: -100,
	}

	want.FaultDescription = new(uint16(0))
	t.Run("success", func(t *testing.T) {
		conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x72, response},
			propertyExchange{[]uint8{0x9F}, 0x72, []Property{
				testProperty(0x9F, 9, 0x80, 0x88, 0x89, 0x9F, 0xD0, 0xE0, 0xE3, 0xE7, 0xE8),
			}},
			propertyExchange{[]uint8{0x80, 0x88, 0x89, 0xD0}, 0x72, []Property{
				testProperty(0x80, 0x30),
				testProperty(0x88, 0x42),
				testProperty(0x89, 0, 0),
				testProperty(0xD0, 2),
			}},
		)
		got, err := NewMultipleInputPCSClient(conn).Get(t.Context(), "127.0.0.1", eoj)
		if err != nil || got == nil || !reflect.DeepEqual(got, &want) {
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

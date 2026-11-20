package echonetlite

import (
	"slices"
	"testing"
)

func TestPVPowerGenerationGet(t *testing.T) {
	eoj := EOJ{0x02, 0x79, 0x01}
	epcs := []uint8{0xE0, 0xE1}
	response := []Property{
		testProperty(0xE0, 0xFF, 0xFD),
		testProperty(0xE1, 0, 0, 0, 200),
	}
	want := PVPowerGeneration{
		InstantaneousElectricPowerGeneration: 65_533, CumulativeElectricEnergyOfGeneration: 200,
	}

	t.Run("success", func(t *testing.T) {
		conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x72, response})
		got, err := NewPVPowerGenerationClient(conn).Get(t.Context(), "127.0.0.1", eoj)
		if err != nil || got == nil || *got != want {
			t.Fatalf("Get = (%+v, %v), want %+v", got, err, want)
		}
	})

	for _, tc := range []struct {
		name     string
		index    int
		property Property
	}{
		{"short power", 0, testProperty(0xE0, 0)},
		{"power overflow", 0, testProperty(0xE0, 0xFF, 0xFE)},
		{"short energy", 1, testProperty(0xE1, 0, 0, 0)},
		{"energy above max", 1, testProperty(0xE1, 0x3B, 0x9A, 0xCA, 0x00)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			props := slices.Clone(response)
			props[tc.index] = tc.property
			conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x72, props})
			if got, err := NewPVPowerGenerationClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}

	t.Run("missing required property", func(t *testing.T) {
		props := slices.Clone(response)
		props[len(props)-1].EDT = nil
		conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x52, props})
		if got, err := NewPVPowerGenerationClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
			t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
		}
	})
}

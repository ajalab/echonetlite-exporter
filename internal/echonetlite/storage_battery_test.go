package echonetlite

import (
	"slices"
	"testing"
)

func TestStorageBatteryGet(t *testing.T) {
	eoj := EOJ{0x02, 0x7D, 0x01}
	// The group is matched by EPC membership, independent of request ordering.
	epcs := []uint8{0xA9, 0xA8, 0xA5, 0xA4}
	response := []Property{
		testProperty(0xA4, 0, 0, 0, 100),
		testProperty(0xA5, 0, 0, 0, 200),
		testProperty(0xA8, 0, 0, 1, 0x2C),
		testProperty(0xA9, 0x3B, 0x9A, 0xC9, 0xFF),
	}
	want := StorageBattery{
		ACChargeableElectricEnergy: 100, ACDischargeableElectricEnergy: 200,
		ACCumulativeChargingElectricEnergy: 300, ACCumulativeDischargingElectricEnergy: 999_999_999,
	}

	t.Run("success", func(t *testing.T) {
		conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x72, response})
		got, err := NewStorageBatteryClient(conn).Get(t.Context(), "127.0.0.1", eoj)
		if err != nil || got == nil || *got != want {
			t.Fatalf("Get = (%+v, %v), want %+v", got, err, want)
		}
	})

	for _, tc := range []struct {
		name     string
		index    int
		property Property
	}{
		{"short chargeable energy", 0, testProperty(0xA4, 0, 0, 100)},
		{"dischargeable energy above max", 1, testProperty(0xA5, 0x3B, 0x9A, 0xCA, 0x00)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			props := slices.Clone(response)
			props[tc.index] = tc.property
			conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x72, props})
			if got, err := NewStorageBatteryClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}

	t.Run("missing required property", func(t *testing.T) {
		props := slices.Clone(response)
		props[len(props)-1].EDT = nil
		conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x52, props})
		if got, err := NewStorageBatteryClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
			t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
		}
	})
}

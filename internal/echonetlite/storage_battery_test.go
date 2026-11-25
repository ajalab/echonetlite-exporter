package echonetlite

import (
	"reflect"
	"slices"
	"testing"
)

func TestStorageBatteryGet(t *testing.T) {
	eoj := EOJ{0x02, 0x7D, 0x01}
	protocol := propertyExchange{[]uint8{0x82, 0x9F}, 0x72, []Property{
		testProperty(0x82, 0, 0, 'H', 0),
		testProperty(0x9F, 27, 0x25, 0x24, 0x05, 0x24, 0x64, 0x64, 0x00, 0x00, 0x15, 0x14, 0x34, 0x24, 0x00, 0x00, 0x00, 0x12),
	}}
	status := propertyExchange{[]uint8{0x80, 0x88, 0xCF, 0xDA, 0xE4}, 0x72, []Property{
		testProperty(0x80, 0x30),
		testProperty(0x88, 0x42),
		testProperty(0xCF, 0x41),
		testProperty(0xDA, 0x46),
		testProperty(0xE4, 85),
	}}
	energyEPCs := []uint8{0xA4, 0xA5, 0xA8, 0xA9, 0xAA, 0xAB, 0xDB}
	response := []Property{
		testProperty(0xA4, 0, 0, 0, 100),
		testProperty(0xA5, 0, 0, 0, 200),
		testProperty(0xA8, 0, 0, 1, 0x2C),
		testProperty(0xA9, 0x3B, 0x9A, 0xC9, 0xFF),
		testProperty(0xAA, 0, 0, 0, 0),
		testProperty(0xAB, 0, 0, 0, 11),
		testProperty(0xDB, 1),
	}
	want := StorageBattery{
		SuperClass:                                SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		ACEffectiveChargingCapacity:               new(uint32(1)),
		ACEffectiveDischargingCapacity:            new(uint32(2)),
		ACChargeableCapacity:                      new(uint32(3)),
		ACDischargeableCapacity:                   new(uint32(4)),
		ACTargetDischargingElectricEnergy:         new(uint32(11)),
		MinimumAndMaximumChargingElectricPower:    [2]*uint32{new(uint32(10)), nil},
		MinimumAndMaximumDischargingElectricPower: [2]*uint32{new(uint32(10)), new(uint32(30))},
		MinimumAndMaximumChargingCurrent:          [2]*uint16{nil, new(uint16(345))},
		ActualOperationMode:                       new(uint8(0x41)),
		OperationMode:                             new(uint8(0x46)),
		PowerSystemInterconnectionStatus:          new(uint8(0x01)),
		RatedElectricEnergy:                       new(uint32(90)),
		RatedCapacity:                             new(uint16(12)),
		RemainingCapacity3:                        new(uint8(85)),
		ACChargeableElectricEnergy:                100, ACDischargeableElectricEnergy: 200,
		ACCumulativeChargingElectricEnergy: 300, ACCumulativeDischargingElectricEnergy: 999_999_999,
	}

	t.Run("success", func(t *testing.T) {
		conn := newPropertyTestConnection(t, eoj, protocol, status,
			propertyExchange{energyEPCs, 0x72, response},
			propertyExchange{[]uint8{0xA0, 0xA1, 0xA2, 0xA3, 0xC8, 0xC9}, 0x72, []Property{
				testProperty(0xA0, 0, 0, 0, 1),
				testProperty(0xA1, 0, 0, 0, 2),
				testProperty(0xA2, 0, 0, 0, 3),
				testProperty(0xA3, 0, 0, 0, 4),
				testProperty(0xC8, 0, 0, 0, 10, 0xFF, 0xFF, 0xFF, 0xFF),
				testProperty(0xC9, 0, 0, 0, 10, 0, 0, 0, 30),
			}},
			propertyExchange{[]uint8{0xD0, 0xD1}, 0x52, []Property{testProperty(0xD0, 0, 0, 0, 90)}},
			propertyExchange{[]uint8{0xCA}, 0x72, []Property{testProperty(0xCA, 0xFF, 0xFF, 0x01, 0x59)}},
			propertyExchange{[]uint8{0xD3}, 0x72, []Property{testProperty(0xD3, 0x80, 0, 0, 0)}},
			propertyExchange{[]uint8{0xD4}, 0x72, []Property{testProperty(0xD4, 0x7F, 0xFF)}},
			propertyExchange{[]uint8{0xD5}, 0x72, []Property{testProperty(0xD5, 0x80, 0)}},
			propertyExchange{[]uint8{0xE5}, 0x52, []Property{testProperty(0xE5)}},
			propertyExchange{[]uint8{0xD1}, 0x72, []Property{testProperty(0xD1, 0, 12)}},
		)
		got, err := NewStorageBatteryClient(conn).Get(t.Context(), "127.0.0.1", eoj)
		if err != nil || got == nil || !reflect.DeepEqual(got, &want) {
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
			conn := newPropertyTestConnection(t, eoj, protocol, status, propertyExchange{energyEPCs, 0x72, props})
			if got, err := NewStorageBatteryClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}

	t.Run("missing required property", func(t *testing.T) {
		props := slices.Clone(response)
		props[len(props)-1].EDT = nil
		conn := newPropertyTestConnection(t, eoj, protocol, status, propertyExchange{energyEPCs, 0x52, props})
		if got, err := NewStorageBatteryClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
			t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
		}
	})
}

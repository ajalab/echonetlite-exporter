package echonetlite

import (
	"reflect"
	"testing"
)

func TestPowerDistributionBoardMeteringGet(t *testing.T) {
	eoj := EOJ{0x02, 0x87, 0x01}
	conn := newPropertyTestConnection(t, eoj,
		propertyExchange{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 4, 0x9F, 0xB3, 0xB7, 0xC2)}},
		propertyExchange{[]uint8{0xC2}, 0x72, []Property{testProperty(0xC2, 0x01)}},
		propertyExchange{[]uint8{0xB3}, 0x72, []Property{testProperty(0xB3, 2, 2, 0, 0, 0, 5, 0xFF, 0xFF, 0xFF, 0xFE)}},
		propertyExchange{[]uint8{0xB7}, 0x72, []Property{testProperty(0xB7, 4, 1, 0xFF, 0xFF, 0xFF, 0x9C)}},
	)
	got, err := NewPowerDistributionBoardMeteringClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	want := &PowerDistributionBoardMetering{
		UnitForCumulativeElectricEnergy: 0.1,
		CumulativeElectricEnergyListSimplex: CumulativeElectricEnergyListSimplex{
			StartChannel: 2, ElectricEnergy: []*uint32{new(uint32(5)), nil},
		},
		InstantaneousElectricPowerListSimplex: InstantaneousElectricPowerListSimplex{
			StartChannel: 4, InstantaneousElectricPower: []*int32{new(int32(-100))},
		},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Get = (%+v, %v), want %+v", got, err, want)
	}
}

func TestPowerDistributionBoardMeteringGetOptionalLists(t *testing.T) {
	eoj := EOJ{0x02, 0x87, 0x01}
	conn := newPropertyTestConnection(t, eoj,
		propertyExchange{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 3, 0x9F, 0xB3, 0xC2)}},
		propertyExchange{[]uint8{0xC2}, 0x72, []Property{testProperty(0xC2, 0x00)}},
		propertyExchange{[]uint8{0xB3}, 0x72, []Property{testProperty(0xB3, 0xFD, 0xFD, 0xFF, 0xFF, 0xFF, 0xFE)}},
	)
	got, err := NewPowerDistributionBoardMeteringClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	want := &PowerDistributionBoardMetering{UnitForCumulativeElectricEnergy: 1}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Get = (%+v, %v), want %+v", got, err, want)
	}
}

func TestPowerDistributionBoardMeteringGetRejectsInvalidLists(t *testing.T) {
	for _, tc := range []struct {
		name string
		edt  []byte
	}{
		{"short list", []byte{3, 2, 0, 0, 0, 10}},
		{"invalid start channel", []byte{0, 1, 0, 0, 0, 10}},
		{"channel range out of bounds", []byte{252, 2, 0, 0, 0, 10, 0, 0, 0, 20}},
		{"trailing bytes", []byte{3, 1, 0, 0, 0, 10, 0}},
		{"invalid energy", []byte{3, 1, 0x05, 0xF5, 0xE1, 0x00}},
		{"unset range with a value", []byte{0xFD, 0xFD, 0, 0, 0, 10}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0x87, 0x01}
			conn := newPropertyTestConnection(t, eoj,
				propertyExchange{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 3, 0x9F, 0xB3, 0xC2)}},
				propertyExchange{[]uint8{0xC2}, 0x72, []Property{testProperty(0xC2, 0x00)}},
				propertyExchange{[]uint8{0xB3}, 0x72, []Property{testProperty(0xB3, tc.edt...)}},
			)
			if got, err := NewPowerDistributionBoardMeteringClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}
}

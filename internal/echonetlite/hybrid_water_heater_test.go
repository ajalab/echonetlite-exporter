package echonetlite

import (
	"testing"
)

func TestHybridWaterHeaterGetWithBitmapPropertyMap(t *testing.T) {
	eoj := EOJ{0x02, 0xA6, 0x01}
	// The bitmap includes B2, B3, B8, C3, E1, and E2.
	bitmap := []byte{16, 0x40, 0x40, 0x48, 0x18, 0x40, 0x40, 0x40, 0x40,
		0x08, 0x40, 0x40, 0x40, 0x40, 0x00, 0x00, 0x02}
	conn := newPropertyTestConnection(t, eoj,
		propertyExchange{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, bitmap...)}},
		propertyExchange{[]uint8{0xB8, 0xB2, 0xC3}, 0x52, []Property{testProperty(0xB8, 0x42), testProperty(0xB2, 0x41), testProperty(0xC3)}},
		propertyExchange{[]uint8{0xB3}, 0x72, []Property{testProperty(0xB3, 0x42)}},
		propertyExchange{[]uint8{0xE1}, 0x72, []Property{testProperty(0xE1, 0, 150)}},
		propertyExchange{[]uint8{0xE2}, 0x52, []Property{testProperty(0xE2)}},
	)
	got, err := NewHybridWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil {
		t.Fatal(err)
	}
	if got.LinkageModeForSolarPowerGeneration != 0x42 || got.RemainingWater == nil || *got.RemainingWater != 150 {
		t.Fatalf("expected linkage mode 0x42 and 150 L: %+v", got)
	}
	if got.WaterHeatingStatus == nil || !*got.WaterHeatingStatus || got.HeaterStatus == nil || *got.HeaterStatus {
		t.Fatalf("expected valid optional heating states: %+v", got)
	}
	if got.HotWaterSupplyStatus != nil || got.TankCapacity != nil {
		t.Fatalf("expected missing optional values to be absent: %+v", got)
	}
}

func TestHybridWaterHeaterGetRejectsInvalidRequiredProperties(t *testing.T) {
	for _, tc := range []struct {
		name      string
		exchanges []propertyExchange
	}{
		{
			name: "invalid mode",
			exchanges: []propertyExchange{
				{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 2, 0xB8, 0x9F)}},
				{[]uint8{0xB8}, 0x72, []Property{testProperty(0xB8, 0x45)}},
			},
		},
		{
			name: "missing mode with a valid optional value",
			exchanges: []propertyExchange{
				{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 3, 0xB2, 0xB8, 0x9F)}},
				{[]uint8{0xB8, 0xB2}, 0x52, []Property{testProperty(0xB8), testProperty(0xB2, 0x41)}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0xA6, 0x01}
			conn := newPropertyTestConnection(t, eoj, tc.exchanges...)
			if got, err := NewHybridWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want no reading and an error", got, err)
			}
		})
	}
}

func TestHybridWaterHeaterGetRejectsInvalidOptionalProperties(t *testing.T) {
	for _, tc := range []struct {
		name string
		epc  uint8
		edt  []byte
	}{
		{"grouped state", 0xB2, []byte{0x43}},
		{"volume length", 0xE1, []byte{150}},
		{"volume range", 0xE2, []byte{0xFF, 0xFE}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0xA6, 0x01}
			group := propertyExchange{[]uint8{0xB8}, 0x72, []Property{testProperty(0xB8, 0x42)}}
			if tc.epc == 0xB2 {
				group.request = append(group.request, tc.epc)
				group.reply = append(group.reply, testProperty(tc.epc, tc.edt...))
			}
			exchanges := []propertyExchange{
				{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 3, 0x9F, 0xB8, tc.epc)}},
				group,
			}
			if tc.epc != 0xB2 {
				exchanges = append(exchanges, propertyExchange{[]uint8{tc.epc}, 0x72, []Property{testProperty(tc.epc, tc.edt...)}})
			}
			conn := newPropertyTestConnection(t, eoj, exchanges...)
			if got, err := NewHybridWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}
}

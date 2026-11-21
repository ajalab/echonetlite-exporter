package echonetlite

import (
	"testing"
)

func TestInstantaneousWaterHeaterGetUsesAIFPropertyGroup(t *testing.T) {
	eoj := EOJ{0x02, 0x72, 0x01}
	conn := newPropertyTestConnection(t, eoj,
		propertyExchange{
			[]uint8{0x9F}, 0x72,
			[]Property{testProperty(0x9F, 9, 0x9F, 0xD0, 0xD1, 0xE1, 0xE2, 0xE3, 0xEA, 0xEB, 0xEF)},
		},
		propertyExchange{
			[]uint8{0xD0, 0xE2, 0xE3}, 0x72,
			[]Property{testProperty(0xD0, 0x41), testProperty(0xE2, 0x42), testProperty(0xE3, 0x42)},
		},
		propertyExchange{[]uint8{0xD1}, 0x72, []Property{testProperty(0xD1, 0)}},
		propertyExchange{[]uint8{0xE1}, 0x52, []Property{testProperty(0xE1)}},
		propertyExchange{[]uint8{0xEA}, 0x52, []Property{testProperty(0xEA)}},
		propertyExchange{[]uint8{0xEB}, 0x72, []Property{testProperty(0xEB, 0x41)}},
		propertyExchange{[]uint8{0xEF}, 0x72, []Property{testProperty(0xEF, 0x43)}},
	)
	got, err := NewInstantaneousWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil {
		t.Fatal(err)
	}
	if !got.HotWaterHeatingStatus || got.BathWaterHeatingStatus {
		t.Fatalf("unexpected mandatory heating states: %+v", got)
	}
	if got.AutomaticBathOperation == nil || *got.AutomaticBathOperation ||
		got.TargetSuppliedWaterTemperature == nil || *got.TargetSuppliedWaterTemperature != 0 ||
		got.KitchenHotWaterSupplyStatus == nil || !*got.KitchenHotWaterSupplyStatus ||
		got.BathOperationStatus == nil || *got.BathOperationStatus != 0x43 {
		t.Fatalf("expected valid decoded optional values: %+v", got)
	}
	if got.TargetBathWaterTemperature != nil || got.ShowerHotWaterSupplyStatus != nil {
		t.Fatalf("expected missing optional values to be absent: %+v", got)
	}
}

func TestInstantaneousWaterHeaterGetSkipsUnsupportedOptionalProperties(t *testing.T) {
	eoj := EOJ{0x02, 0x72, 0x01}
	conn := newPropertyTestConnection(t, eoj,
		propertyExchange{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 3, 0x9F, 0xD0, 0xE2)}},
		propertyExchange{[]uint8{0xD0, 0xE2}, 0x72, []Property{testProperty(0xD0, 0x41), testProperty(0xE2, 0x42)}},
	)
	got, err := NewInstantaneousWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil || got.AutomaticBathOperation != nil {
		t.Fatalf("Get = (%+v, %v), want a reading without automaticBathOperation", got, err)
	}
}

func TestInstantaneousWaterHeaterGetRejectsInvalidRequiredStatus(t *testing.T) {
	for _, tc := range []struct {
		name string
		esv  uint8
		bath Property
	}{
		{"missing heating status", 0x52, testProperty(0xE2)},
		{"invalid heating status", 0x72, testProperty(0xE2, 0x43)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0x72, 0x01}
			conn := newPropertyTestConnection(t, eoj,
				propertyExchange{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 3, 0x9F, 0xD0, 0xE2)}},
				propertyExchange{[]uint8{0xD0, 0xE2}, tc.esv, []Property{testProperty(0xD0, 0x41), tc.bath}},
			)
			if got, err := NewInstantaneousWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}
}

func TestInstantaneousWaterHeaterGetRejectsInvalidOptionalProperties(t *testing.T) {
	for _, tc := range []struct {
		name string
		epc  uint8
		edt  []byte
	}{
		{"grouped state", 0xE3, []byte{0x43}},
		{"state length", 0xEA, []byte{0x41, 0x42}},
		{"temperature range", 0xD1, []byte{101}},
		{"bath operation code", 0xEF, []byte{0x44}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0x72, 0x01}
			group := propertyExchange{
				[]uint8{0xD0, 0xE2}, 0x72,
				[]Property{testProperty(0xD0, 0x41), testProperty(0xE2, 0x42)},
			}
			if tc.epc == 0xE3 {
				group.request = append(group.request, tc.epc)
				group.reply = append(group.reply, testProperty(tc.epc, tc.edt...))
			}
			exchanges := []propertyExchange{
				{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 4, 0x9F, 0xD0, 0xE2, tc.epc)}},
				group,
			}
			if tc.epc != 0xE3 {
				exchanges = append(exchanges, propertyExchange{[]uint8{tc.epc}, 0x72, []Property{testProperty(tc.epc, tc.edt...)}})
			}
			conn := newPropertyTestConnection(t, eoj, exchanges...)
			if got, err := NewInstantaneousWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}
}

package echonetlite

import (
	"reflect"
	"testing"
)

func TestInstantaneousWaterHeaterGetUsesAIFPropertyGroup(t *testing.T) {
	eoj := EOJ{0x02, 0x72, 0x01}
	conn := newPropertyTestConnection(t, eoj,
		propertyExchange{
			[]uint8{0x9F}, 0x72,
			[]Property{testProperty(0x9F, 11, 0x9F, 0xD0, 0xD1, 0xDB, 0xE1, 0xE2, 0xE3, 0xE7, 0xEA, 0xEB, 0xEF)},
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
		propertyExchange{[]uint8{0x80, 0x88}, 0x72, []Property{testProperty(0x80, 0x30), testProperty(0x88, 0x42)}},
		propertyExchange{[]uint8{0xDB}, 0x72, []Property{testProperty(0xDB, 0xFF, 0xFF)}},
		propertyExchange{[]uint8{0xE7}, 0x72, []Property{testProperty(0xE7, 227)}},
	)
	got, err := NewInstantaneousWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil {
		t.Fatal(err)
	}
	if got.HotWaterHeatingStatus != 0x41 || got.BathWaterHeatingStatus != 0x42 {
		t.Fatalf("unexpected mandatory heating states: %+v", got)
	}
	if got.AutomaticBathOperation == nil || *got.AutomaticBathOperation != 0x42 ||
		got.TargetSuppliedWaterTemperature == nil || *got.TargetSuppliedWaterTemperature != 0 ||
		got.KitchenHotWaterSupplyStatus == nil || *got.KitchenHotWaterSupplyStatus != 0x41 ||
		got.BathOperationStatus == nil || *got.BathOperationStatus != 0x43 {
		t.Fatalf("expected valid decoded optional values: %+v", got)
	}
	if got.TargetBathWaterTemperature != nil || got.ShowerHotWaterSupplyStatus != nil || got.RemainingAutomaticOperationTime != nil {
		t.Fatalf("expected missing optional values to be absent: %+v", got)
	}
	if got.BathWaterVolume1 == nil || *got.BathWaterVolume1 != 227 {
		t.Fatalf("expected the bath volume in liters: %+v", got)
	}
}

func TestInstantaneousWaterHeaterGetSkipsUnsupportedOptionalProperties(t *testing.T) {
	eoj := EOJ{0x02, 0x72, 0x01}
	conn := newPropertyTestConnection(t, eoj,
		propertyExchange{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 3, 0x9F, 0xD0, 0xE2)}},
		propertyExchange{[]uint8{0xD0, 0xE2}, 0x72, []Property{testProperty(0xD0, 0x41), testProperty(0xE2, 0x42)}},
		propertyExchange{[]uint8{0x80, 0x88}, 0x72, []Property{testProperty(0x80, 0x30), testProperty(0x88, 0x42)}},
	)
	got, err := NewInstantaneousWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil || got.AutomaticBathOperation != nil {
		t.Fatalf("Get = (%+v, %v), want a reading without automaticBathOperation", got, err)
	}
}

func TestInstantaneousWaterHeaterGetSkipsOutOfRangeBathVolumes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		volume1 Property
		volume3 Property
		want1   *uint8
		want3   *uint16
	}{
		{"volume 1 underflow", testProperty(0xE7, 0xFE), testProperty(0xEE, 0x01, 0x2C), nil, new(uint16(300))},
		{"volume 1 overflow", testProperty(0xE7, 0xFF), testProperty(0xEE, 0x01, 0x2C), nil, new(uint16(300))},
		{"volume 3 underflow", testProperty(0xE7, 253), testProperty(0xEE, 0xFF, 0xFE), new(uint8(253)), nil},
		{"volume 3 overflow", testProperty(0xE7, 253), testProperty(0xEE, 0xFF, 0xFF), new(uint8(253)), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0x72, 0x01}
			conn := newPropertyTestConnection(t, eoj,
				propertyExchange{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 5, 0x9F, 0xD0, 0xE2, 0xE7, 0xEE)}},
				propertyExchange{[]uint8{0xD0, 0xE2}, 0x72, []Property{testProperty(0xD0, 0x41), testProperty(0xE2, 0x42)}},
				propertyExchange{[]uint8{0x80, 0x88}, 0x72, []Property{testProperty(0x80, 0x30), testProperty(0x88, 0x42)}},
				propertyExchange{[]uint8{0xE7}, 0x72, []Property{tc.volume1}},
				propertyExchange{[]uint8{0xEE}, 0x72, []Property{tc.volume3}},
			)
			got, err := NewInstantaneousWaterHeaterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
			if err != nil {
				t.Fatal(err)
			}
			if got.HotWaterHeatingStatus != 0x41 || got.BathWaterHeatingStatus != 0x42 ||
				!reflect.DeepEqual(got.BathWaterVolume1, tc.want1) || !reflect.DeepEqual(got.BathWaterVolume3, tc.want3) {
				t.Fatalf("expected heating states and only the in-range bath volume: %+v", got)
			}
		})
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
		{"bath volume 1 length", 0xE7, []byte{0xFF, 0xFF}},
		{"bath volume 3 length", 0xEE, []byte{0xFF}},
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
			if tc.epc == 0xE7 || tc.epc == 0xEE {
				exchanges = append(exchanges, propertyExchange{[]uint8{0x80, 0x88}, 0x72, []Property{testProperty(0x80, 0x30), testProperty(0x88, 0x42)}})
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

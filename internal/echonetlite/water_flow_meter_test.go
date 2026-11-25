package echonetlite

import (
	"testing"
)

func TestWaterFlowMeterGet(t *testing.T) {
	eoj := EOJ{0x02, 0x81, 0x01}
	// A short Get_SNA can omit an optional EPC even when the map advertises it.
	group := propertyExchange{[]uint8{0x80, 0x88, 0xE3}, 0x72, []Property{
		testProperty(0x80, 0x30),
		testProperty(0x88, 0x42),
		testProperty(0xE3, 0x41),
	}}
	group.request = append(group.request, 0xE4)
	group.esv = esvGetSNA
	conn := newPropertyTestConnection(t, eoj,
		propertyExchange{[]uint8{0xE0, 0xE1}, 0x72, []Property{
			testProperty(0xE0, 0x3B, 0x9A, 0xC9, 0xFF),
			testProperty(0xE1, 0x06),
		}},
		propertyExchange{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 8, 0x80, 0x88, 0x9F, 0xE0, 0xE1, 0xE3, 0xE4, 0xE6)}},
		group,
		propertyExchange{[]uint8{0xE4}, 0x72, []Property{testProperty(0xE4, 0xFF, 0xFF, 0xFF, 0xFF)}},
	)
	got, err := NewWaterFlowMeterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil || got.CumulativeFlowingWater != 999_999_999 || got.UnitForCumulativeFlowingWater != 0.000001 {
		t.Fatalf("Get = (%+v, %v), want flow 999999999 and unit 0.000001", got, err)
	}
	if got.AbnormalDetection == nil || *got.AbnormalDetection != 0x41 ||
		got.SecurityDataInformation == nil || *got.SecurityDataInformation != 0xFFFFFFFF {
		t.Fatalf("expected raw water meter status and bitmask: %+v", got)
	}
}

func TestWaterFlowMeterGetRejectsInvalidRequiredProperties(t *testing.T) {
	volume, unit := testProperty(0xE0, 0, 0, 0, 1), testProperty(0xE1, 0x06)
	for _, tc := range []struct {
		name  string
		esv   uint8
		reply []Property
	}{
		{"missing unit", 0x52, []Property{volume, testProperty(0xE1)}},
		{"unknown unit", 0x72, []Property{volume, testProperty(0xE1, 0x07)}},
		{"invalid volume", 0x72, []Property{testProperty(0xE0, 0x3B, 0x9A, 0xCA, 0x00), unit}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0x81, 0x01}
			conn := newPropertyTestConnection(t, eoj, propertyExchange{[]uint8{0xE0, 0xE1}, tc.esv, tc.reply})
			if got, err := NewWaterFlowMeterClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}
}

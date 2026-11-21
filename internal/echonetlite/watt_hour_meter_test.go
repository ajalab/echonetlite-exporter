package echonetlite

import "testing"

func TestWattHourMeterGet(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response []Property
		esv      uint8
		wantUnit float64
		wantErr  bool
	}{
		{"tenths of a kWh", []Property{testProperty(0xE2, 0x01), testProperty(0xE0, 0, 0, 0, 100)}, 0x72, 0.1, false},
		{"hundredths of a kWh", []Property{testProperty(0xE0, 0, 0, 0, 100), testProperty(0xE2, 0x02)}, 0x72, 0.01, false},
		{"missing unit", []Property{testProperty(0xE0, 0, 0, 0, 100), testProperty(0xE2)}, 0x52, 0, true},
		{"unknown unit", []Property{testProperty(0xE0, 0, 0, 0, 100), testProperty(0xE2, 0x00)}, 0x72, 0, true},
		{"short EDT", []Property{testProperty(0xE0, 0, 100), testProperty(0xE2, 0x01)}, 0x72, 0, true},
		{"out of range", []Property{testProperty(0xE0, 0xFF, 0xFF, 0xFF, 0xFE), testProperty(0xE2, 0x01)}, 0x72, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0x80, 0x01}
			conn := newPropertyTestConnection(t, eoj, propertyExchange{[]uint8{0xE0, 0xE2}, tc.esv, tc.response})
			got, err := NewWattHourMeterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
			if tc.wantErr {
				if got != nil || err == nil {
					t.Fatalf("Get = (%+v, %v), want no reading and an error", got, err)
				}
				return
			}
			if err != nil || got.CumulativeElectricEnergy != 100 || got.CumulativeAmountsOfElectricEnergyUnit != tc.wantUnit {
				t.Fatalf("Get = (%+v, %v), want energy 100 and unit %v", got, err, tc.wantUnit)
			}
		})
	}
}

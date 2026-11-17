package echonetlite

import (
	"net"
	"slices"
	"strings"
	"testing"
)

func TestWattHourMeterGet(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response []Property
		wantUnit float64
		wantErr  string
	}{
		{"tenths of a kWh", []Property{testProperty(0xE2, 0x01), testProperty(0xE0, 0, 0, 0, 100)}, 0.1, ""},
		{"hundredths of a kWh", []Property{testProperty(0xE0, 0, 0, 0, 100), testProperty(0xE2, 0x02)}, 0.01, ""},
		{"missing unit", []Property{testProperty(0xE0, 0, 0, 0, 100)}, 0, "missing required EPC 0xE2"},
		{"empty unit", []Property{testProperty(0xE0, 0, 0, 0, 100), testProperty(0xE2)}, 0, "missing required EPC 0xE2"},
		{"unknown unit", []Property{testProperty(0xE0, 0, 0, 0, 100), testProperty(0xE2, 0x00)}, 0, "invalid cumulativeAmountsOfElectricEnergyUnit"},
		{"duplicate EPC", []Property{testProperty(0xE0, 0, 0, 0, 100), testProperty(0xE0, 0, 0, 0, 100)}, 0, "duplicate EPC 0xE0"},
		{"short EDT", []Property{testProperty(0xE0, 0, 100), testProperty(0xE2, 0x01)}, 0, "expected 4 bytes"},
		{"out of range", []Property{testProperty(0xE0, 0xFF, 0xFF, 0xFF, 0xFE), testProperty(0xE2, 0x01)}, 0, "exceeds"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eoj := EOJ{0x02, 0x80, 0x01}
			conn := newPropertyTestConnection(t, eoj, propertyExchange{[]uint8{0xE0, 0xE2}, 0x72, tc.response})
			got, err := NewWattHourMeterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
			if tc.wantErr != "" {
				if got != nil || err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("Get = (%+v, %v), want no reading and an error containing %q", got, err, tc.wantErr)
				}
				return
			}
			if err != nil || got.CumulativeElectricEnergy != 100 || got.CumulativeAmountsOfElectricEnergyUnit != tc.wantUnit {
				t.Fatalf("Get = (%+v, %v), want energy 100 and unit %v", got, err, tc.wantUnit)
			}
		})
	}
}

func TestWaterFlowMeterGet(t *testing.T) {
	eoj := EOJ{0x02, 0x81, 0x01}
	conn := newPropertyTestConnection(t, eoj, propertyExchange{
		[]uint8{0xE0, 0xE1}, 0x72,
		[]Property{testProperty(0xE0, 0x3B, 0x9A, 0xC9, 0xFF), testProperty(0xE1, 0x06)},
	})
	got, err := NewWaterFlowMeterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil || got.CumulativeFlowingWater != 999_999_999 || got.UnitForCumulativeFlowingWater != 0.000001 {
		t.Fatalf("Get = (%+v, %v), want flow 999999999 and unit 0.000001", got, err)
	}
}

func TestGasMeterGet(t *testing.T) {
	eoj := EOJ{0x02, 0x82, 0x01}
	conn := newPropertyTestConnection(t, eoj, propertyExchange{
		[]uint8{0xE0}, 0x72, []Property{testProperty(0xE0, 0, 0, 0, 100)},
	})
	got, err := NewGasMeterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil || got.CumulativeGasConsumption != 100 {
		t.Fatalf("Get = (%+v, %v), want consumption count 100", got, err)
	}
}

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
		propertyExchange{[]uint8{0xE1}, 0x72, []Property{testProperty(0xE1, 0xFF)}},
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
		t.Fatalf("expected invalid or missing optional values to be absent: %+v", got)
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
		propertyExchange{[]uint8{0xE2}, 0x72, []Property{testProperty(0xE2, 0xFF, 0xFF)}},
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
		t.Fatalf("expected missing or invalid optional values to be absent: %+v", got)
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
		{
			name: "invalid property map",
			exchanges: []propertyExchange{
				{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, 3, 0xB8, 0x9F)}},
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

type propertyExchange struct {
	request []uint8
	esv     uint8
	reply   []Property
}

func testProperty(epc uint8, edt ...byte) Property {
	return Property{EPC: epc, EDT: edt}
}

func newPropertyTestConnection(t *testing.T, eoj EOJ, exchanges ...propertyExchange) *Connection {
	t.Helper()
	conn := &Connection{}
	t.Cleanup(func() {
		if len(exchanges) != 0 {
			t.Errorf("%d expected requests were not sent", len(exchanges))
		}
	})
	conn.conn = &stubPacketConn{writeToFunc: func(packet []byte, addr net.Addr) (int, error) {
		request, err := Deserialize(packet)
		if err != nil {
			t.Fatalf("Deserialize: %v", err)
		}
		if request.DEOJ != eoj || request.ESV != 0x62 || len(exchanges) == 0 {
			t.Fatalf("unexpected request: %+v", request)
		}
		var epcs []uint8
		for _, property := range request.Properties {
			epcs = append(epcs, property.EPC)
		}
		exchange := exchanges[0]
		exchanges = exchanges[1:]
		if !slices.Equal(epcs, exchange.request) {
			t.Fatalf("requested EPCs %X, want %X", epcs, exchange.request)
		}
		channel, ok := conn.pending.Load(request.TID)
		if !ok {
			t.Fatalf("no pending request for TID %d", request.TID)
		}
		channel.(chan FrameAddr) <- FrameAddr{Addr: addr, Frame: &Frame{
			TID: request.TID, SEOJ: request.DEOJ, DEOJ: request.SEOJ, ESV: exchange.esv, Properties: exchange.reply,
		}}
		return len(packet), nil
	}}
	return conn
}

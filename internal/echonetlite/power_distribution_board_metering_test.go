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
		propertyExchange{[]uint8{0xB7}, 0x72, []Property{testProperty(0xB7, 4, 2, 0xFF, 0xFF, 0xFF, 0x9C, 0x80, 0, 0, 0)}},
		propertyExchange{[]uint8{0x80, 0x88, 0xC0, 0xC1}, 0x72, []Property{
			testProperty(0x80, 0x30),
			testProperty(0x88, 0x42),
			testProperty(0xC0, 0xFF, 0xFF, 0xFF, 0xFE),
			testProperty(0xC1, 0, 0, 0, 1),
		}},
	)
	got, err := NewPowerDistributionBoardMeteringClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	want := &PowerDistributionBoardMetering{
		SuperClass:                               SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		ReverseDirectionCumulativeElectricEnergy: new(uint32(1)),
		UnitForCumulativeElectricEnergy:          0.1,
		CumulativeElectricEnergyListSimplex: CumulativeElectricEnergyListSimplex{
			StartChannel: 2, ElectricEnergy: []*uint32{new(uint32(5)), nil},
		},
		InstantaneousElectricPowerListSimplex: InstantaneousElectricPowerListSimplex{
			StartChannel: 4, InstantaneousElectricPower: []*int32{new(int32(-100)), nil},
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
		propertyExchange{[]uint8{0x80, 0x88, 0xC0, 0xC1}, 0x72, []Property{
			testProperty(0x80, 0x30),
			testProperty(0x88, 0x42),
			testProperty(0xC0, 0xFF, 0xFF, 0xFF, 0xFE),
			testProperty(0xC1, 0, 0, 0, 1),
		}},
	)
	got, err := NewPowerDistributionBoardMeteringClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	want := &PowerDistributionBoardMetering{
		SuperClass:                               SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		UnitForCumulativeElectricEnergy:          1,
		ReverseDirectionCumulativeElectricEnergy: new(uint32(1)),
	}
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

func TestPowerDistributionBoardMeteringGetCompoundMeasurements(t *testing.T) {
	eoj := EOJ{0x02, 0x87, 0x01}
	conn := newPropertyTestConnection(t, eoj,
		propertyExchange{[]uint8{0x9F}, 0x72, []Property{
			testProperty(0x9F, 15, 0x80, 0x88, 0x9F, 0xB1, 0xB5, 0xBA, 0xBC, 0xBE, 0xC0, 0xC1, 0xC2, 0xC6, 0xC7, 0xC8, 0xEF),
		}},
		propertyExchange{[]uint8{0xC2}, 0x72, []Property{testProperty(0xC2, 0x02)}},
		propertyExchange{[]uint8{0x80, 0x88, 0xB1, 0xC0, 0xC1, 0xC6, 0xC7, 0xC8, 0xEF}, 0x72, []Property{
			testProperty(0x80, 0x30),
			testProperty(0x88, 0x42),
			testProperty(0xB1, 0xFD),
			testProperty(0xC0, 0, 0, 0, 0),
			testProperty(0xC1, 0xFF, 0xFF, 0xFF, 0xFE),
			testProperty(0xC6, 0x80, 0, 0, 0),
			testProperty(0xC7, 0xFF, 0xCE, 0x7F, 0xFF),
			testProperty(0xC8, 0x03, 0xE8, 0xFF, 0xFF),
			testProperty(0xEF, 0, 0, 0, 6, 0x80, 0x00, 0x00, 0x32),
		}},
		propertyExchange{[]uint8{0xB5}, 0x72, []Property{testProperty(0xB5, 0xFD, 0xFD, 0x7F, 0xFE, 0x7F, 0xFE)}},
		propertyExchange{[]uint8{0xBA}, 0x72, []Property{testProperty(0xBA, 4, 1, 0, 0, 0, 7, 0xFF, 0xFF, 0xFF, 0xFE)}},
		propertyExchange{[]uint8{0xBC}, 0x72, []Property{testProperty(0xBC, 9, 1, 0x7F, 0xFE, 0x00, 0x64)}},
		propertyExchange{[]uint8{0xBE}, 0x72, []Property{testProperty(0xBE, 9, 2, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF)}},
	)
	got, err := NewPowerDistributionBoardMeteringClient(conn).Get(t.Context(), "127.0.0.1", eoj)
	if err != nil {
		t.Fatal(err)
	}
	want := &PowerDistributionBoardMetering{
		SuperClass:                              SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		UnitForCumulativeElectricEnergy:         0.01,
		NormalDirectionCumulativeElectricEnergy: new(uint32(0)),
		InstantaneousCurrent:                    &CurrentPhases{RPhase: new(int16(-50))},
		InstantaneousVoltage:                    &LineVoltage{RS: new(uint16(1000))},
		MeasurementChannel32:                    &MeasurementChannel{ElectricEnergy: new(uint32(6)), CurrentTPhase: new(int16(50))},
		InstantaneousCurrentListSimplex:         &InstantaneousCurrentList{},
		CumulativeElectricEnergyListDuplex: &CumulativeElectricEnergyListDuplex{
			StartChannel: 4, ElectricEnergy: []DirectionalEnergy{{NormalDirectionElectricEnergy: new(uint32(7))}},
		},
		InstantaneousCurrentListDuplex: &InstantaneousCurrentList{
			StartChannel: 9, InstantaneousCurrent: []CurrentPhases{{TPhase: new(int16(100))}},
		},
		InstantaneousElectricPowerListDuplex: &InstantaneousElectricPowerListDuplex{
			StartChannel: 9, InstantaneousElectricPower: []*int32{nil, new(int32(-1))},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Get = %+v, want %+v", got, want)
	}
}

func TestDecodeLineVoltagePreservesSecondLinePair(t *testing.T) {
	got, err := decodeLineVoltage([]byte{0xFF, 0xFF, 0x00, 0x00})
	want := LineVoltage{ST: new(uint16(0))}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("decodeLineVoltage = (%+v, %v), want %+v", got, err, want)
	}
}

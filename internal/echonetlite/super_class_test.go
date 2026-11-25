package echonetlite

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestSuperClassEPCs(t *testing.T) {
	var supported [256]bool
	supported[EPCCurrentLimit] = true
	supported[EPCFaultDescription] = true
	supported[0xD0] = true

	required, optional := superClassEPCs(supported)
	if !slices.Equal(required, []uint8{EPCOperationStatus, EPCFaultStatus}) ||
		!slices.Equal(optional, []uint8{EPCCurrentLimit, EPCFaultDescription}) {
		t.Fatalf("default EPCs: required %X, optional %X", required, optional)
	}

	required, optional = superClassEPCs(supported, EPCFaultDescription, 0xD0)
	if !slices.Equal(required, []uint8{EPCOperationStatus, EPCFaultStatus, EPCFaultDescription, 0xD0}) ||
		!slices.Equal(optional, []uint8{EPCCurrentLimit}) {
		t.Fatalf("EPCs with a class requirement: required %X, optional %X", required, optional)
	}
}

func TestDecodeSuperClass(t *testing.T) {
	reading, err := decodeSuperClass(map[uint8][]byte{
		EPCOperationStatus:  {0x31},
		EPCFaultStatus:      {0x42},
		EPCFaultDescription: {0, 0},
		EPCHourMeter:        {0x43, 0, 0, 0, 2},
	})
	want := SuperClass{
		OperationStatus: new(uint8(0x31)), FaultStatus: new(uint8(0x42)),
		FaultDescription: new(uint16(0)), HourMeter: new(uint64(7200)),
	}
	if err != nil || !reflect.DeepEqual(reading, want) {
		t.Fatalf("decodeSuperClass = (%+v, %v), want %+v", reading, err, want)
	}
}

func TestDecodeSuperClassPowerRangeCodes(t *testing.T) {
	for _, tc := range []struct {
		name string
		low  byte
	}{
		{"underflow", 0xFE},
		{"overflow", 0xFF},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reading, err := decodeSuperClass(map[uint8][]byte{
				EPCOperationStatus:                       {0x30},
				EPCFaultStatus:                           {0x42},
				EPCInstantaneousElectricPowerConsumption: {0xFF, tc.low},
				EPCPowerLimit:                            {0xFF, tc.low},
				EPCHourMeter:                             {0x41, 0xFF, 0xFF, 0xFF, 0xFF},
			})
			want := SuperClass{
				OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42)),
				PowerLimit: new(0xFF00 | uint16(tc.low)),
			}
			if err != nil || !reflect.DeepEqual(reading, want) {
				t.Fatalf("decodeSuperClass = (%+v, %v), want %+v", reading, err, want)
			}
		})
	}
}

func TestDecodeSuperClassRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name string
		epc  uint8
		edt  []byte
		want string
	}{
		{"missing required status", EPCOperationStatus, nil, "0x80"},
		{"invalid fault status", EPCFaultStatus, []byte{0xFF}, "0x88"},
		{"malformed optional fault description", EPCFaultDescription, []byte{0}, "0x89"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[uint8][]byte{EPCOperationStatus: {0x30}, EPCFaultStatus: {0x42}}
			values[tc.epc] = tc.edt
			if _, err := decodeSuperClass(values); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("decodeSuperClass error = %v, want EPC %s", err, tc.want)
			}
		})
	}
}

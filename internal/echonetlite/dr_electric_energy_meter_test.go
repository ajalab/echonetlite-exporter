package echonetlite

import (
	"reflect"
	"testing"
)

func TestDRElectricEnergyMeterGet(t *testing.T) {
	eoj := EOJ{0x02, 0x8E, 0x01}
	input := testProperty(0xE0, 0, 0, 0, 100)
	output := testProperty(0xE2, 0, 0, 0, 200)
	unit := testProperty(0xD4, 0x01)
	for _, tc := range []struct {
		name      string
		readable  []byte
		exchanges []propertyExchange
		want      *DRElectricEnergyMeter
	}{
		{
			name:     "input only with optional NoData",
			readable: []byte{0x9F, 0xD4, 0xE0, 0xE4, 0xE9, 0xEA},
			exchanges: []propertyExchange{
				{[]uint8{0xD4}, 0x72, []Property{unit}},
				{[]uint8{0xE0, 0xE4, 0xE9, 0xEA}, 0x72, []Property{
					input, testProperty(0xE4, 0xFF, 0xFF, 0xFF, 0xFE),
					testProperty(0xE9, 0x7F, 0xFF, 0xFF, 0xFE), testProperty(0xEA, 0x7F, 0xFF, 0xFF, 0xFE),
				}},
			},
			want: &DRElectricEnergyMeter{
				CumulativeAmountsOfElectricEnergyUnit: 0.1,
				ACInputCumulativeElectricEnergy:       new(uint32(100)),
			},
		},
		{
			name:     "output only with optional partial SNA",
			readable: []byte{0x9F, 0xD4, 0xE2, 0xE9, 0xEA},
			exchanges: []propertyExchange{
				{[]uint8{0xD4}, 0x72, []Property{unit}},
				{[]uint8{0xE2, 0xE9, 0xEA}, 0x52, []Property{output, testProperty(0xE9, 0xFF, 0xFF, 0xFF, 0x9C), testProperty(0xEA)}},
			},
			want: &DRElectricEnergyMeter{
				CumulativeAmountsOfElectricEnergyUnit: 0.1,
				ACOutputCumulativeElectricEnergy:      new(uint32(200)),
				ACInstantaneousElectricPower:          new(int32(-100)),
			},
		},
		{
			name:     "both directions",
			readable: []byte{0x9F, 0xD4, 0xE0, 0xE2},
			exchanges: []propertyExchange{
				{[]uint8{0xD4}, 0x72, []Property{unit}},
				{[]uint8{0xE0, 0xE2}, 0x72, []Property{input, output}},
			},
			want: &DRElectricEnergyMeter{
				CumulativeAmountsOfElectricEnergyUnit: 0.1,
				ACInputCumulativeElectricEnergy:       new(uint32(100)),
				ACOutputCumulativeElectricEnergy:      new(uint32(200)),
			},
		},
		{
			name:     "neither direction",
			readable: []byte{0x9F, 0xD4, 0xEA},
		},
		{
			name:     "missing output from a meter that advertised it",
			readable: []byte{0x9F, 0xD4, 0xE0, 0xE2},
			exchanges: []propertyExchange{
				{[]uint8{0xD4}, 0x72, []Property{unit}},
				{[]uint8{0xE0, 0xE2}, 0x52, []Property{input, testProperty(0xE2)}},
			},
		},
		{
			name:     "invalid unit",
			readable: []byte{0x9F, 0xD4, 0xE0},
			exchanges: []propertyExchange{
				{[]uint8{0xD4}, 0x72, []Property{testProperty(0xD4, 0x05)}},
			},
		},
		{
			name:     "invalid required energy",
			readable: []byte{0x9F, 0xD4, 0xE0},
			exchanges: []propertyExchange{
				{[]uint8{0xD4}, 0x72, []Property{unit}},
				{[]uint8{0xE0}, 0x72, []Property{testProperty(0xE0, 0xFF, 0xFF, 0xFF, 0xFE)}},
			},
		},
		{
			name:     "invalid optional energy",
			readable: []byte{0x9F, 0xD4, 0xE0, 0xE4},
			exchanges: []propertyExchange{
				{[]uint8{0xD4}, 0x72, []Property{unit}},
				{[]uint8{0xE0, 0xE4}, 0x72, []Property{input, testProperty(0xE4, 0x3B, 0x9A, 0xCA, 0x00)}},
			},
		},
		{
			name:     "invalid optional power",
			readable: []byte{0x9F, 0xD4, 0xE0, 0xE9},
			exchanges: []propertyExchange{
				{[]uint8{0xD4}, 0x72, []Property{unit}},
				{[]uint8{0xE0, 0xE9}, 0x72, []Property{input, testProperty(0xE9, 1)}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mapEDT := append([]byte{byte(len(tc.readable))}, tc.readable...)
			exchanges := append([]propertyExchange{{[]uint8{0x9F}, 0x72, []Property{testProperty(0x9F, mapEDT...)}}}, tc.exchanges...)
			conn := newPropertyTestConnection(t, eoj, exchanges...)
			got, err := NewDRElectricEnergyMeterClient(conn).Get(t.Context(), "127.0.0.1", eoj)
			if (err == nil) != (tc.want != nil) || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Get = (%+v, %v), want %+v", got, err, tc.want)
			}
		})
	}
}

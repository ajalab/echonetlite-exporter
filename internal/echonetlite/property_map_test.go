package echonetlite

import "testing"

func TestDecodePropertyMap(t *testing.T) {
	list := []byte{15, 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E}
	bitmap := []byte{16, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0x80} // 0x80–0x8E, 0xFF
	for _, tc := range []struct {
		name string
		edt  []byte
		last int
	}{
		{"list with 15 EPCs", list, 0x8E},
		{"bitmap with 16 EPCs", bitmap, 0xFF},
		{"missing count", nil, 0},
		{"short list", []byte{2, 0x80}, 0},
		{"invalid EPC", []byte{1, 0x7F}, 0},
		{"duplicate EPC", []byte{2, 0x80, 0x80}, 0},
		{"short bitmap", bitmap[:16], 0},
		{"wrong bitmap count", append([]byte{17}, bitmap[1:]...), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodePropertyMap(tc.edt)
			if tc.last == 0 {
				if err == nil {
					t.Fatal("expected invalid property map")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for epc, present := range got {
				want := epc >= 0x80 && epc <= 0x8E || epc == tc.last
				if present != want {
					t.Errorf("EPC %02X present = %v, want %v", epc, present, want)
				}
			}
		})
	}
}

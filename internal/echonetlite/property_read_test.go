package echonetlite

import (
	"reflect"
	"testing"
)

func TestGetResponseValues(t *testing.T) {
	req := newGetRequest(EOJ{0x02, 0x72, 0x01}, []uint8{0xD0, 0xE2, 0xE3})
	a, b, c := testProperty(0xD0, 0x41), testProperty(0xE2, 0x42), testProperty(0xE3, 0x41)

	for _, tc := range []struct {
		name  string
		esv   uint8
		props []Property
		want  map[uint8][]byte
	}{
		{"success", 0x72, []Property{b, a, c}, map[uint8][]byte{0xD0: a.EDT, 0xE2: b.EDT, 0xE3: c.EDT}},
		{"missing EPC", 0x72, []Property{a, b}, nil},
		{"empty EDT", 0x72, []Property{a, b, testProperty(0xE3)}, nil},
		{"duplicate EPC", 0x72, []Property{a, b, b}, nil},
		{"unrequested EPC", 0x72, []Property{a, b, testProperty(0x80, 0x30)}, nil},
		{"wrong ESV", 0x73, []Property{a, b, c}, nil},
		{"partial SNA", 0x52, []Property{a, testProperty(0xE2), c}, map[uint8][]byte{0xD0: a.EDT, 0xE2: nil, 0xE3: c.EDT}},
		{"short SNA prefix", 0x52, []Property{a, b}, map[uint8][]byte{0xD0: a.EDT, 0xE2: b.EDT}},
		{"SNA without failure", 0x52, []Property{a, b, c}, nil},
		{"short SNA with skipped EPC", 0x52, []Property{a, c}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := &Frame{SEOJ: req.DEOJ, DEOJ: req.SEOJ, ESV: tc.esv, Properties: tc.props}
			got, err := getResponseValues(req, res)
			if (err == nil) != (tc.want != nil) || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got (%v, %v), want %v", got, err, tc.want)
			}
		})
	}

	for _, objects := range []struct{ source, destination EOJ }{
		{req.SEOJ, req.SEOJ},
		{req.DEOJ, req.DEOJ},
	} {
		res := &Frame{SEOJ: objects.source, DEOJ: objects.destination, ESV: 0x72, Properties: []Property{a, b, c}}
		if got, err := getResponseValues(req, res); got != nil || err == nil {
			t.Fatalf("wrong response EOJ accepted: %v, %v", got, err)
		}
	}
}

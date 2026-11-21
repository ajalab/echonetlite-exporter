package echonetlite

import (
	"bytes"
	"testing"
)

func TestDeserializeGetFrames(t *testing.T) {
	valid := []byte{0x10, 0x81, 0x12, 0x34, 0x02, 0x80, 1, 0x05, 0xFF, 1, 0x52, 2, 0xE0, 4, 0, 0, 0, 1, 0xE2, 0}
	frame, err := Deserialize(valid)
	if err != nil || !bytes.Equal(frame.Serialize(), valid) {
		t.Fatalf("Get_SNA round trip = (%+v, %v)", frame, err)
	}

	for _, tc := range []struct {
		name   string
		packet []byte
	}{
		{"header too short", valid[:11]},
		{"unknown EHD1", append([]byte{0x00}, valid[1:]...)},
		{"unknown EHD2", append([]byte{0x10, 0x82}, valid[2:]...)},
		{"zero OPC", append(bytes.Clone(valid[:11]), 0)},
		{"cut off property header", valid[:len(valid)-1]},
		{"cut off EDT", valid[:17]},
		{"extra data", append(bytes.Clone(valid), 0)},
		{"OPC too small", append(append(bytes.Clone(valid[:11]), 1), valid[12:]...)},
		{"OPC too large", append(append(bytes.Clone(valid[:11]), 3), valid[12:]...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, err := Deserialize(tc.packet); got != nil || err == nil {
				t.Fatalf("Deserialize = (%+v, %v), want an error and no frame", got, err)
			}
		})
	}
}

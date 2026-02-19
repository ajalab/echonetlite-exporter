package echonetlite

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type Property struct {
	EPC uint8
	EDT []byte
}

type Frame struct {
	TID        uint16
	SEOJ       EOJ
	DEOJ       EOJ
	ESV        uint8
	Properties []Property
}

func (f *Frame) Serialize() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x10, 0x81})
	binary.Write(buf, binary.BigEndian, f.TID)
	buf.Write(f.SEOJ[:])
	buf.Write(f.DEOJ[:])
	buf.WriteByte(f.ESV)
	buf.WriteByte(uint8(len(f.Properties)))

	for _, p := range f.Properties {
		buf.WriteByte(p.EPC)
		buf.WriteByte(uint8(len(p.EDT)))
		buf.Write(p.EDT)
	}
	return buf.Bytes()
}

func Deserialize(data []byte) (*Frame, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("packet too short")
	}
	f := &Frame{
		TID: binary.BigEndian.Uint16(data[2:4]),
		ESV: data[10],
	}
	copy(f.SEOJ[:], data[4:7])
	copy(f.DEOJ[:], data[7:10])

	opc := int(data[11])
	offset := 12
	for i := 0; i < opc; i++ {
		if offset+2 > len(data) {
			break
		}
		epc := data[offset]
		pdc := int(data[offset+1])
		offset += 2
		if offset+pdc > len(data) {
			break
		}
		f.Properties = append(f.Properties, Property{
			EPC: epc,
			EDT: data[offset : offset+pdc],
		})
		offset += pdc
	}
	return f, nil
}

type FrameAddr struct {
	Frame *Frame
	Addr  net.Addr
}

package echonetlite

import (
	"net"
	"slices"
	"testing"
	"time"
)

type stubPacketConn struct {
	writeToFunc func([]byte, net.Addr) (int, error)
}

func (s *stubPacketConn) ReadFrom(_ []byte) (int, net.Addr, error) {
	return 0, nil, net.ErrClosed
}

func (s *stubPacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	if s.writeToFunc != nil {
		return s.writeToFunc(p, addr)
	}
	return len(p), nil
}

func (s *stubPacketConn) Close() error {
	return nil
}

func (s *stubPacketConn) LocalAddr() net.Addr {
	return &net.UDPAddr{}
}

func (s *stubPacketConn) SetDeadline(_ time.Time) error {
	return nil
}

func (s *stubPacketConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (s *stubPacketConn) SetWriteDeadline(_ time.Time) error {
	return nil
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
	conn := &Connection{responseWait: time.Second}
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
		if request.SEOJ != (EOJ{0x05, 0xFF, 0x01}) || request.DEOJ != eoj || request.ESV != 0x62 || addr.String() != "127.0.0.1:3610" || len(exchanges) == 0 {
			t.Fatalf("unexpected request: %+v", request)
		}
		var epcs []uint8
		for _, property := range request.Properties {
			if len(property.EDT) != 0 {
				t.Fatalf("Get must have an empty EDT: %+v", property)
			}
			epcs = append(epcs, property.EPC)
		}
		exchange := exchanges[0]
		exchanges = exchanges[1:]
		expectedEPCs := slices.Clone(exchange.request)
		slices.Sort(epcs)
		slices.Sort(expectedEPCs)
		if !slices.Equal(epcs, expectedEPCs) {
			t.Fatalf("requested EPCs %X, want %X", epcs, expectedEPCs)
		}
		reply := Frame{
			TID: request.TID, SEOJ: request.DEOJ, DEOJ: request.SEOJ, ESV: exchange.esv, Properties: exchange.reply,
		}
		conn.dispatchResponse(&reply, addr)
		return len(packet), nil
	}}
	return conn
}

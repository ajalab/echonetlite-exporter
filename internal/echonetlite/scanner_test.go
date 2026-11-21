package echonetlite

import (
	"context"
	"net"
	"reflect"
	"testing"
	"testing/synctest"
	"time"
)

func TestScanDiscoversNodesFromValidResponses(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		conn := &Connection{}
		conn.conn = &stubPacketConn{writeToFunc: func(packet []byte, addr net.Addr) (int, error) {
			req, err := Deserialize(packet)
			if err != nil {
				t.Fatal(err)
			}
			wantReq := Frame{
				TID: req.TID, SEOJ: EOJ{0x05, 0xFF, 1}, DEOJ: EOJ{0x0E, 0xF0, 1}, ESV: 0x62,
				Properties: []Property{testProperty(0x80), testProperty(0xD6)},
			}
			if addr.String() != "224.0.23.0:3610" || !reflect.DeepEqual(packet, wantReq.Serialize()) {
				t.Fatalf("unexpected multicast request to %v: %+v", addr, req)
			}

			reply := Frame{
				TID: req.TID, SEOJ: req.DEOJ, DEOJ: req.SEOJ, ESV: 0x72,
				Properties: []Property{testProperty(0x80, 0x30), testProperty(0xD6, 2, 0x02, 0x80, 1, 0x02, 0x80, 2)},
			}
			conn.dispatchResponse(&reply, &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 3610})
			bad := reply
			bad.Properties = []Property{testProperty(0x80, 0x30), testProperty(0xD6, 1)}
			conn.dispatchResponse(&bad, &net.UDPAddr{IP: net.ParseIP("192.0.2.11"), Port: 3610})
			badStatus := reply
			badStatus.Properties = []Property{testProperty(0x80, 0x32), testProperty(0xD6, 0)}
			conn.dispatchResponse(&badStatus, &net.UDPAddr{IP: net.ParseIP("192.0.2.12"), Port: 3610})
			return len(packet), nil
		}}
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		got, err := NewNodeProfileClient(conn).Scan(ctx)
		want := []Node{{
			Host: "192.0.2.10",
			NodeProfile: NodeProfile{
				OperatingStatus: true, SelfNodeInstanceListS: []EOJ{{0x02, 0x80, 1}, {0x02, 0x80, 2}},
			},
		}}
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("Scan = (%+v, %v), want %+v", got, err, want)
		}
	})
}

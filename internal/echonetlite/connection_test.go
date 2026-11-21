package echonetlite

import (
	"context"
	"errors"
	"net"
	"testing"
	"testing/synctest"
	"time"
)

func TestUnicastMatchesSourceAndTID(t *testing.T) {
	req := newGetRequest(EOJ{0x02, 0x82, 0x01}, []uint8{0xE0})
	conn := &Connection{responseWait: time.Second}
	conn.conn = &stubPacketConn{writeToFunc: func(packet []byte, addr net.Addr) (int, error) {
		sent, err := Deserialize(packet)
		if err != nil {
			t.Fatal(err)
		}
		reply := Frame{TID: sent.TID, SEOJ: sent.DEOJ, DEOJ: sent.SEOJ, ESV: 0x72}
		wrongTID := reply
		wrongTID.TID++
		conn.dispatchResponse(&wrongTID, addr)
		wrongSource := reply
		conn.dispatchResponse(&wrongSource, &net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 3610})

		reply.Properties = []Property{testProperty(0xE0, 0, 0, 0, 1)}
		source := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 49152}
		conn.dispatchResponse(&reply, source)
		conn.dispatchResponse(&reply, source) // Duplicate while the first response is buffered.
		return len(packet), nil
	}}

	res, err := conn.unicast(t.Context(), "127.0.0.1", req, 0)
	if err != nil || len(res.Properties) != 1 {
		t.Fatalf("unicast got (%+v, %v), want response from the requested address and TID", res, err)
	}
}

func TestUnicastSerializesPerNodeAndWaitsForAIFMinimum(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		type sentRequest struct {
			frame *Frame
			addr  net.Addr
		}
		sent := make(chan sentRequest, 3)
		conn := &Connection{responseWait: 10 * time.Second}
		conn.conn = &stubPacketConn{writeToFunc: func(packet []byte, addr net.Addr) (int, error) {
			frame, err := Deserialize(packet)
			if err != nil {
				return 0, err
			}
			sent <- sentRequest{frame, addr}
			return len(packet), nil
		}}
		start := func(host string, eoj EOJ, minimumWait time.Duration) <-chan error {
			result := make(chan error, 1)
			go func() {
				_, err := conn.unicast(ctx, host, newGetRequest(eoj, []uint8{0x80}), minimumWait)
				result <- err
			}()
			return result
		}
		first := start("127.0.0.1", EOJ{0x02, 0x8E, 1}, 20*time.Second)
		<-sent
		second := start("127.0.0.1", EOJ{0x02, 0x79, 1}, 20*time.Second)
		other := start("127.0.0.2", EOJ{0x02, 0x82, 1}, 0)
		synctest.Wait()
		request := <-sent
		if request.addr.String() != "127.0.0.2:3610" {
			t.Fatalf("queued same-node request was sent early: %v", request.addr)
		}
		conn.dispatchResponse(&Frame{TID: request.frame.TID}, request.addr)
		if err := <-other; err != nil {
			t.Fatal(err)
		}

		time.Sleep(19 * time.Second)
		synctest.Wait()
		select {
		case req := <-sent:
			t.Fatalf("same-node request was sent before the AIF timeout: %+v", req)
		default:
		}
		time.Sleep(time.Second)
		if err := <-first; !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("first request error = %v, want deadline exceeded", err)
		}
		request = <-sent
		if request.frame.DEOJ != (EOJ{0x02, 0x79, 1}) {
			t.Fatalf("unexpected next request: %+v", request)
		}
		cancel()
		if err := <-second; !errors.Is(err, context.Canceled) {
			t.Fatalf("second request error = %v, want cancellation", err)
		}
	})
}

func TestRegisterRequestSkipsTIDInUseAfterWraparound(t *testing.T) {
	conn := &Connection{lastTID: 0xFFFF}
	first := &pendingRequest{}
	if tid, err := conn.registerRequest(first); tid != 0 || err != nil {
		t.Fatalf("first TID = %d, error = %v", tid, err)
	}
	conn.lastTID = 0xFFFF
	if tid, err := conn.registerRequest(&pendingRequest{}); tid != 1 || err != nil {
		t.Fatalf("second TID = %d, error = %v", tid, err)
	}
	if pending, _ := conn.pending.Load(uint16(0)); pending != first {
		t.Fatalf("existing request was overwritten: %v", pending)
	}
}

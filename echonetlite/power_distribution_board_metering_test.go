package echonetlite

import (
	"context"
	"net"
	"testing"
)

func TestParseInstantaneousPowerList(t *testing.T) {
	edts := []byte{
		0x01, 0x02, // start channel 1, range 2
		0x00, 0x00, 0x00, 0x64, // 100
		0xFF, 0xFF, 0xFF, 0x9C, // -100
	}

	ibls := parseInstantaneousPowerListSimplex(edts)
	if ibls.StartChannel != 1 {
		t.Fatalf("expected start channel 1, got %d", ibls.StartChannel)
	}
	if ibls.Range != 2 {
		t.Fatalf("expected range 2, got %d", ibls.Range)
	}
	if len(ibls.InstantaneousElectricPower) != 2 {
		t.Fatalf("expected 2 values, got %d", len(ibls.InstantaneousElectricPower))
	}
	if ibls.InstantaneousElectricPower[0] != 100 {
		t.Fatalf("expected first value 100, got %d", ibls.InstantaneousElectricPower[0])
	}
	if ibls.InstantaneousElectricPower[1] != -100 {
		t.Fatalf("expected second value -100, got %d", ibls.InstantaneousElectricPower[1])
	}
}

func TestPowerDistributionBoardMeteringGetUsesGivenHostAndEOJ(t *testing.T) {
	conn := &Connection{}
	stub := &stubPacketConn{}
	conn.conn = stub

	var gotFrame *Frame
	var gotAddr *net.UDPAddr
	stub.writeToFunc = func(packet []byte, addr net.Addr) (int, error) {
		frame, err := Deserialize(packet)
		if err != nil {
			t.Fatalf("Deserialize failed: %v", err)
		}
		gotFrame = frame
		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok {
			t.Fatalf("expected UDP addr, got %T", addr)
		}
		gotAddr = udpAddr

		res := &Frame{
			TID:        frame.TID,
			SEOJ:       frame.DEOJ,
			DEOJ:       frame.SEOJ,
			ESV:        0x72,
			Properties: []Property{},
		}
		ch, ok := conn.pending.Load(frame.TID)
		if !ok {
			t.Fatalf("pending channel not found for tid=%d", frame.TID)
		}
		ch.(chan Response) <- Response{Frame: res, Addr: addr}
		return len(packet), nil
	}

	client := NewPowerDistributionBoardMeteringClient(conn)
	host := "127.0.0.1"
	eoj := EOJ{0x02, 0x87, 0x02}

	if _, err := client.Get(context.Background(), host, eoj); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if gotFrame == nil {
		t.Fatal("no frame was sent")
	}
	if gotFrame.DEOJ != eoj {
		t.Fatalf("expected DEOJ %v, got %v", eoj, gotFrame.DEOJ)
	}
	if gotAddr == nil {
		t.Fatal("no destination address captured")
	}
	if gotAddr.IP.String() != host {
		t.Fatalf("expected host %s, got %s", host, gotAddr.IP.String())
	}
}

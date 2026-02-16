package echonetlite

import (
	"context"
	"net"
	"testing"
)

func TestParsePVInstantaneousPower(t *testing.T) {
	val, err := parseInstantaneousElectricPowerGeneration([]byte{0x00, 0x64})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
	if _, err := parseInstantaneousElectricPowerGeneration([]byte{0x00}); err == nil {
		t.Fatalf("expected error for invalid length")
	}
}

func TestParsePVCumulativeEnergy(t *testing.T) {
	val, err := parseCumulativeElectricEnergyOfGeneration([]byte{0x00, 0x00, 0x0F, 0xA0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 4000 {
		t.Fatalf("expected 4000, got %d", val)
	}
	if _, err := parseCumulativeElectricEnergyOfGeneration([]byte{0x00, 0x00, 0x0F}); err == nil {
		t.Fatalf("expected error for invalid length")
	}
}

func TestPVPowerGenerationGetUsesGivenHostAndEOJ(t *testing.T) {
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

	pv := NewPVPowerGenerationClient(conn)
	host := "127.0.0.1"
	eoj := EOJ{0x02, 0x79, 0x01}

	if _, err := pv.Get(context.Background(), host, eoj); err != nil {
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

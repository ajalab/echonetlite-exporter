package echonetlite

import (
	"context"
	"net"
	"testing"
)

func TestParseMultipleInputPCSElectricEnergy(t *testing.T) {
	val, err := parseMultipleInputPCSElectricEnergy([]byte{0x00, 0x00, 0x0F, 0xA0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 4000 {
		t.Fatalf("expected 4000, got %d", val)
	}
	if _, err := parseMultipleInputPCSElectricEnergy([]byte{0x00, 0x00, 0x0F}); err == nil {
		t.Fatalf("expected error for invalid length")
	}
}

func TestParseMultipleInputPCSInstantaneousElectricPower(t *testing.T) {
	pos, err := parseMultipleInputPCSInstantaneousElectricPower([]byte{0x00, 0x00, 0x00, 0x64})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos != 100 {
		t.Fatalf("expected 100, got %d", pos)
	}

	neg, err := parseMultipleInputPCSInstantaneousElectricPower([]byte{0xFF, 0xFF, 0xFF, 0x9C})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if neg != -100 {
		t.Fatalf("expected -100, got %d", neg)
	}

	if _, err := parseMultipleInputPCSInstantaneousElectricPower([]byte{0x00, 0x00, 0x00}); err == nil {
		t.Fatalf("expected error for invalid length")
	}
}

func TestMultipleInputPCSGetParsesResponse(t *testing.T) {
	client := newMultipleInputPCSTestClient(t, []Property{
		{EPC: epcMultipleInputPCSNormalDirectionElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0xC8}},
		{EPC: epcMultipleInputPCSReverseDirectionElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
		{EPC: epcMultipleInputPCSInstantaneousElectricPower, EDT: []byte{0xFF, 0xFF, 0xFF, 0x9C}},
	})

	mipcs, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0xA5, 0x01})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if mipcs.NormalDirectionElectricEnergy != 200 {
		t.Fatalf("expected E0=200, got %d", mipcs.NormalDirectionElectricEnergy)
	}
	if mipcs.ReverseDirectionElectricEnergy != 100 {
		t.Fatalf("expected E3=100, got %d", mipcs.ReverseDirectionElectricEnergy)
	}
	if mipcs.InstantaneousElectricPower != -100 {
		t.Fatalf("expected E7=-100, got %d", mipcs.InstantaneousElectricPower)
	}
}

func TestMultipleInputPCSGetFailsOnInvalidLength(t *testing.T) {
	client := newMultipleInputPCSTestClient(t, []Property{
		{EPC: epcMultipleInputPCSNormalDirectionElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
		{EPC: epcMultipleInputPCSReverseDirectionElectricEnergy, EDT: []byte{0x00}},
		{EPC: epcMultipleInputPCSInstantaneousElectricPower, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0xA5, 0x01}); err == nil {
		t.Fatal("expected invalid length error")
	}
}

func TestMultipleInputPCSGetFailsOnUnexpectedEPC(t *testing.T) {
	client := newMultipleInputPCSTestClient(t, []Property{
		{EPC: epcMultipleInputPCSNormalDirectionElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
		{EPC: epcMultipleInputPCSReverseDirectionElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
		{EPC: 0x99, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0xA5, 0x01}); err == nil {
		t.Fatal("expected unexpected EPC error")
	}
}

func TestMultipleInputPCSGetFailsOnUnexpectedPropertyCount(t *testing.T) {
	client := newMultipleInputPCSTestClient(t, []Property{
		{EPC: epcMultipleInputPCSNormalDirectionElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
		{EPC: epcMultipleInputPCSReverseDirectionElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0xA5, 0x01}); err == nil {
		t.Fatal("expected unexpected property count error")
	}
}

func newMultipleInputPCSTestClient(t *testing.T, responseProperties []Property) *MultipleInputPCSClient {
	t.Helper()

	conn := &Connection{}
	stub := &stubPacketConn{}
	conn.conn = stub

	stub.writeToFunc = func(packet []byte, addr net.Addr) (int, error) {
		frame, err := Deserialize(packet)
		if err != nil {
			t.Fatalf("Deserialize failed: %v", err)
		}

		res := &Frame{
			TID:        frame.TID,
			SEOJ:       frame.DEOJ,
			DEOJ:       frame.SEOJ,
			ESV:        0x72,
			Properties: responseProperties,
		}
		ch, ok := conn.pending.Load(frame.TID)
		if !ok {
			t.Fatalf("pending channel not found for tid=%d", frame.TID)
		}
		ch.(chan FrameAddr) <- FrameAddr{Frame: res, Addr: addr}
		return len(packet), nil
	}

	return NewMultipleInputPCSClient(conn)
}

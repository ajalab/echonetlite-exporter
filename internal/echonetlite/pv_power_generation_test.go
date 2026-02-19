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

func TestPVPowerGenerationGetParsesResponse(t *testing.T) {
	client := newPVPowerGenerationTestClient(t, []Property{
		{EPC: epcInstantaneousElectricPowerGeneration, EDT: []byte{0x00, 0x64}},
		{EPC: epcCumulativeElectricEnergyOfGeneration, EDT: []byte{0x00, 0x00, 0x00, 0xC8}},
	})

	pv, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x79, 0x01})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if pv.InstantaneousElectricPowerGeneration != 100 {
		t.Fatalf("expected E0=100, got %d", pv.InstantaneousElectricPowerGeneration)
	}
	if pv.CumulativeElectricEnergyOfGeneration != 200 {
		t.Fatalf("expected E1=200, got %d", pv.CumulativeElectricEnergyOfGeneration)
	}
}

func TestPVPowerGenerationGetFailsOnInvalidLength(t *testing.T) {
	client := newPVPowerGenerationTestClient(t, []Property{
		{EPC: epcInstantaneousElectricPowerGeneration, EDT: []byte{0x00}},
		{EPC: epcCumulativeElectricEnergyOfGeneration, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x79, 0x01}); err == nil {
		t.Fatal("expected invalid length error")
	}
}

func newPVPowerGenerationTestClient(t *testing.T, responseProperties []Property) *PVPowerGenerationClient {
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

	return NewPVPowerGenerationClient(conn)
}

package echonetlite

import (
	"context"
	"net"
	"testing"
)

func TestParseStorageBatteryWh(t *testing.T) {
	val, err := parseStorageBatteryWh([]byte{0x00, 0x00, 0x00, 0x64})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
	if _, err := parseStorageBatteryWh([]byte{0x00, 0x00, 0x64}); err == nil {
		t.Fatal("expected error for invalid length")
	}
}

func TestStorageBatteryGetParsesResponse(t *testing.T) {
	client := newStorageBatteryTestClient(t, []Property{
		{EPC: epcACChargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
		{EPC: epcACDischargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0xC8}},
		{EPC: epcACCumulativeChargingElectricEnergy, EDT: []byte{0x00, 0x00, 0x01, 0x2C}},
		{EPC: epcACCumulativeDischargingElectricEnergy, EDT: []byte{0x00, 0x00, 0x01, 0x90}},
	})

	sb, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x7D, 0x01})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sb.ACChargeableElectricEnergy != 100 {
		t.Fatalf("expected A4=100, got %d", sb.ACChargeableElectricEnergy)
	}
	if sb.ACDischargeableElectricEnergy != 200 {
		t.Fatalf("expected A5=200, got %d", sb.ACDischargeableElectricEnergy)
	}
	if sb.ACCumulativeChargingElectricEnergy != 300 {
		t.Fatalf("expected A8=300, got %d", sb.ACCumulativeChargingElectricEnergy)
	}
	if sb.ACCumulativeDischargingElectricEnergy != 400 {
		t.Fatalf("expected A9=400, got %d", sb.ACCumulativeDischargingElectricEnergy)
	}
}

func TestStorageBatteryGetFailsOnInvalidLength(t *testing.T) {
	client := newStorageBatteryTestClient(t, []Property{
		{EPC: epcACChargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x64}},
		{EPC: epcACDischargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0xC8}},
		{EPC: epcACCumulativeChargingElectricEnergy, EDT: []byte{0x00, 0x00, 0x01, 0x2C}},
		{EPC: epcACCumulativeDischargingElectricEnergy, EDT: []byte{0x00, 0x00, 0x01, 0x90}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x7D, 0x01}); err == nil {
		t.Fatal("expected invalid length error")
	}
}

func newStorageBatteryTestClient(t *testing.T, responseProperties []Property) *StorageBatteryClient {
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

	return NewStorageBatteryClient(conn)
}

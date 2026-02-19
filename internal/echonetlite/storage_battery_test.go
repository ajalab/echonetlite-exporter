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
	client, _, _ := newStorageBatteryTestClient(t, []Property{
		{EPC: epcACChargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
		{EPC: epcACDischargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0xC8}},
	})

	sb, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x7D, 0x01})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sb.ACChargeableElectricEnergyWh != 100 {
		t.Fatalf("expected A4=100, got %d", sb.ACChargeableElectricEnergyWh)
	}
	if sb.ACDischargeableElectricEnergyWh != 200 {
		t.Fatalf("expected A5=200, got %d", sb.ACDischargeableElectricEnergyWh)
	}
}

func TestStorageBatteryGetFailsWhenA4Missing(t *testing.T) {
	client, _, _ := newStorageBatteryTestClient(t, []Property{
		{EPC: epcACDischargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x7D, 0x01}); err == nil {
		t.Fatal("expected error when A4 is missing")
	}
}

func TestStorageBatteryGetFailsWhenA5Missing(t *testing.T) {
	client, _, _ := newStorageBatteryTestClient(t, []Property{
		{EPC: epcACChargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x7D, 0x01}); err == nil {
		t.Fatal("expected error when A5 is missing")
	}
}

func TestStorageBatteryGetFailsOnInvalidLength(t *testing.T) {
	client, _, _ := newStorageBatteryTestClient(t, []Property{
		{EPC: epcACChargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x64}},
		{EPC: epcACDischargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0xC8}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x7D, 0x01}); err == nil {
		t.Fatal("expected invalid length error")
	}
}

func TestStorageBatteryGetUsesGivenHostAndEOJ(t *testing.T) {
	client, gotFrame, gotAddr := newStorageBatteryTestClient(t, []Property{
		{EPC: epcACChargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
		{EPC: epcACDischargeableElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0xC8}},
	})

	host := "127.0.0.1"
	eoj := EOJ{0x02, 0x7D, 0x01}
	if _, err := client.Get(context.Background(), host, eoj); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if *gotFrame == nil {
		t.Fatal("no frame was sent")
	}
	if (*gotFrame).DEOJ != eoj {
		t.Fatalf("expected DEOJ %v, got %v", eoj, (*gotFrame).DEOJ)
	}
	if len((*gotFrame).Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len((*gotFrame).Properties))
	}
	if (*gotFrame).Properties[0].EPC != epcACChargeableElectricEnergy {
		t.Fatalf("expected first EPC 0xA4, got 0x%X", (*gotFrame).Properties[0].EPC)
	}
	if (*gotFrame).Properties[1].EPC != epcACDischargeableElectricEnergy {
		t.Fatalf("expected second EPC 0xA5, got 0x%X", (*gotFrame).Properties[1].EPC)
	}
	if *gotAddr == nil {
		t.Fatal("no destination address captured")
	}
	if (*gotAddr).IP.String() != host {
		t.Fatalf("expected host %s, got %s", host, (*gotAddr).IP.String())
	}
}

func newStorageBatteryTestClient(t *testing.T, responseProperties []Property) (*StorageBatteryClient, **Frame, **net.UDPAddr) {
	t.Helper()

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
			Properties: responseProperties,
		}
		ch, ok := conn.pending.Load(frame.TID)
		if !ok {
			t.Fatalf("pending channel not found for tid=%d", frame.TID)
		}
		ch.(chan FrameAddr) <- FrameAddr{Frame: res, Addr: addr}
		return len(packet), nil
	}

	return NewStorageBatteryClient(conn), &gotFrame, &gotAddr
}

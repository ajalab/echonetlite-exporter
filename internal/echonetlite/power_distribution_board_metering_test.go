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

func TestPowerDistributionBoardMeteringGetParsesResponse(t *testing.T) {
	client := newPowerDistributionBoardMeteringTestClient(t, []Property{
		{EPC: epcCumulativeElectricEnergyListSimplex, EDT: []byte{0x01, 0x01, 0x00, 0x00, 0x00, 0xC8}},
		{EPC: epcInstantaneousElectricPowerListSimplex, EDT: []byte{0x01, 0x01, 0xFF, 0xFF, 0xFF, 0x9C}},
		{EPC: epcUnitForCumulativeElectricEnergy, EDT: []byte{0x02}},
	})

	m, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x87, 0x02})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if m.CumulativeElectricEnergyListSimplex.StartChannel != 1 {
		t.Fatalf("expected B3 start channel 1, got %d", m.CumulativeElectricEnergyListSimplex.StartChannel)
	}
	if len(m.CumulativeElectricEnergyListSimplex.ElectricEnergy) != 1 {
		t.Fatalf("expected one B3 value, got %d", len(m.CumulativeElectricEnergyListSimplex.ElectricEnergy))
	}
	if m.CumulativeElectricEnergyListSimplex.ElectricEnergy[0] != 200 {
		t.Fatalf("expected B3 first value 200, got %d", m.CumulativeElectricEnergyListSimplex.ElectricEnergy[0])
	}
	if len(m.InstantaneousElectricPowerListSimplex.InstantaneousElectricPower) != 1 {
		t.Fatalf("expected one B7 value, got %d", len(m.InstantaneousElectricPowerListSimplex.InstantaneousElectricPower))
	}
	if m.InstantaneousElectricPowerListSimplex.InstantaneousElectricPower[0] != -100 {
		t.Fatalf("expected B7 first value -100, got %d", m.InstantaneousElectricPowerListSimplex.InstantaneousElectricPower[0])
	}
	if m.UnitForCumulativeEnergy != 0.01 {
		t.Fatalf("expected C2 unit 0.01, got %f", m.UnitForCumulativeEnergy)
	}
}

func newPowerDistributionBoardMeteringTestClient(t *testing.T, responseProperties []Property) *PowerDistributionBoardMeteringClient {
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

	return NewPowerDistributionBoardMeteringClient(conn)
}

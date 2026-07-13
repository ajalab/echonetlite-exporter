package echonetlite

import (
	"context"
	"net"
	"testing"
)

func TestParseDRElectricEnergyMeterUnit(t *testing.T) {
	if got := parseDRElectricEnergyMeterUnit([]byte{0x00}); got != 1.0 {
		t.Fatalf("expected 1.0, got %f", got)
	}
	if got := parseDRElectricEnergyMeterUnit([]byte{0x02}); got != 0.01 {
		t.Fatalf("expected 0.01, got %f", got)
	}
	if got := parseDRElectricEnergyMeterUnit([]byte{0x0A}); got != 10.0 {
		t.Fatalf("expected 10.0, got %f", got)
	}
	if got := parseDRElectricEnergyMeterUnit([]byte{0xFF}); got != 1.0 {
		t.Fatalf("expected default 1.0, got %f", got)
	}
}

func TestParseDRElectricEnergyMeterCumulative(t *testing.T) {
	val, err := parseDRElectricEnergyMeterCumulative([]byte{0x00, 0x00, 0x00, 0x64})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
	if _, err := parseDRElectricEnergyMeterCumulative([]byte{0x00}); err == nil {
		t.Fatal("expected invalid length error")
	}
}

func TestParseDRElectricEnergyMeterCumulativeWithNoData(t *testing.T) {
	val, err := parseDRElectricEnergyMeterCumulativeWithNoData([]byte{0x00, 0x00, 0x00, 0x64})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
	if _, err := parseDRElectricEnergyMeterCumulativeWithNoData([]byte{0xFF, 0xFF, 0xFF, 0xFE}); err == nil {
		t.Fatal("expected no data error")
	}
}

func TestParseDRElectricEnergyMeterInstantaneousPowerWithNoData(t *testing.T) {
	val, err := parseDRElectricEnergyMeterInstantaneousPowerWithNoData([]byte{0x00, 0x00, 0x00, 0x64})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
	val, err = parseDRElectricEnergyMeterInstantaneousPowerWithNoData([]byte{0xFF, 0xFF, 0xFF, 0x9C})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != -100 {
		t.Fatalf("expected -100, got %d", val)
	}
	if _, err := parseDRElectricEnergyMeterInstantaneousPowerWithNoData([]byte{0x7F, 0xFF, 0xFF, 0xFE}); err == nil {
		t.Fatal("expected no data error")
	}
}

func TestDRElectricEnergyMeterGetParsesResponse(t *testing.T) {
	client := newDRElectricEnergyMeterTestClient(t, []Property{
		{EPC: epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit, EDT: []byte{0x02}},
		{EPC: epcDRElectricEnergyMeterACInputCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x0A}},
		{EPC: epcDRElectricEnergyMeterACOutputCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x14}},
		{EPC: epcDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x1E}},
		{EPC: epcDRElectricEnergyMeterACInstantaneousElectricPower, EDT: []byte{0xFF, 0xFF, 0xFF, 0x9C}},
		{EPC: epcDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower, EDT: []byte{0x00, 0x00, 0x00, 0x64}},
	})

	m, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x8E, 0x01})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if m.CumulativeAmountsOfElectricEnergyUnit != 0.01 {
		t.Fatalf("expected D4=0.01, got %f", m.CumulativeAmountsOfElectricEnergyUnit)
	}
	if m.ACInputCumulativeElectricEnergy != 10 {
		t.Fatalf("expected E0=10, got %d", m.ACInputCumulativeElectricEnergy)
	}
	if m.ACOutputCumulativeElectricEnergy != 20 {
		t.Fatalf("expected E2=20, got %d", m.ACOutputCumulativeElectricEnergy)
	}
	if m.IndependentOperationCumulativeElectricEnergy != 30 {
		t.Fatalf("expected E4=30, got %d", m.IndependentOperationCumulativeElectricEnergy)
	}
	if m.ACInstantaneousElectricPower != -100 {
		t.Fatalf("expected E9=-100, got %d", m.ACInstantaneousElectricPower)
	}
	if m.IndependentOperationInstantaneousElectricPower != 100 {
		t.Fatalf("expected EA=100, got %d", m.IndependentOperationInstantaneousElectricPower)
	}
}

func TestDRElectricEnergyMeterGetFailsOnMissingPropertyCount(t *testing.T) {
	client := newDRElectricEnergyMeterTestClient(t, []Property{
		{EPC: epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit, EDT: []byte{0x00}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x8E, 0x01}); err == nil {
		t.Fatal("expected property count error")
	}
}

func TestDRElectricEnergyMeterGetFailsOnInvalidLength(t *testing.T) {
	client := newDRElectricEnergyMeterTestClient(t, []Property{
		{EPC: epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit, EDT: []byte{0x00}},
		{EPC: epcDRElectricEnergyMeterACInputCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: epcDRElectricEnergyMeterACOutputCumulativeElectricEnergy, EDT: []byte{0x00}},
		{EPC: epcDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: epcDRElectricEnergyMeterACInstantaneousElectricPower, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: epcDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x8E, 0x01}); err == nil {
		t.Fatal("expected invalid length error")
	}
}

func TestDRElectricEnergyMeterGetFailsOnUnexpectedEPC(t *testing.T) {
	client := newDRElectricEnergyMeterTestClient(t, []Property{
		{EPC: epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit, EDT: []byte{0x00}},
		{EPC: epcDRElectricEnergyMeterACInputCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: epcDRElectricEnergyMeterACOutputCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: epcDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: 0xEE, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: epcDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x8E, 0x01}); err == nil {
		t.Fatal("expected unexpected EPC error")
	}
}

func TestDRElectricEnergyMeterGetFailsOnNoDataSentinel(t *testing.T) {
	client := newDRElectricEnergyMeterTestClient(t, []Property{
		{EPC: epcDRElectricEnergyMeterCumulativeAmountsOfElectricEnergyUnit, EDT: []byte{0x00}},
		{EPC: epcDRElectricEnergyMeterACInputCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: epcDRElectricEnergyMeterACOutputCumulativeElectricEnergy, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: epcDRElectricEnergyMeterIndependentOperationCumulativeElectricEnergy, EDT: []byte{0xFF, 0xFF, 0xFF, 0xFE}},
		{EPC: epcDRElectricEnergyMeterACInstantaneousElectricPower, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
		{EPC: epcDRElectricEnergyMeterIndependentOperationInstantaneousElectricPower, EDT: []byte{0x00, 0x00, 0x00, 0x01}},
	})
	if _, err := client.Get(context.Background(), "127.0.0.1", EOJ{0x02, 0x8E, 0x01}); err == nil {
		t.Fatal("expected no data error")
	}
}

func newDRElectricEnergyMeterTestClient(t *testing.T, responseProperties []Property) *DRElectricEnergyMeterClient {
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

	return NewDRElectricEnergyMeterClient(conn)
}

package echonetlite

import "testing"

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

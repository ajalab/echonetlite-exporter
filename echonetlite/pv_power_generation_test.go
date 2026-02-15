package echonetlite

import "testing"

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

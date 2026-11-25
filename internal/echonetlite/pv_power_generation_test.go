package echonetlite

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestPVPowerGenerationGet(t *testing.T) {
	eoj := EOJ{0x02, 0x79, 0x01}
	epcs := []uint8{0xE0, 0xE1}
	response := []Property{
		testProperty(0xE0, 0xFF, 0xFD),
		testProperty(0xE1, 0, 0, 0, 200),
	}
	want := PVPowerGeneration{
		SuperClass:                           SuperClass{OperationStatus: new(uint8(0x30)), FaultStatus: new(uint8(0x42))},
		OutputPowerControl1:                  new(uint8(80)),
		ApprovedCapacity:                     new(uint16(1000)),
		PowerSystemInterconnectionStatus:     new(uint8(0x03)),
		OutputPowerRestraintStatus:           new(uint8(0x45)),
		CumulativeElectricEnergySold:         new(uint32(45)),
		InstantaneousElectricPowerGeneration: 65_533, CumulativeElectricEnergyOfGeneration: 200,
	}

	want.FaultDescription = new(uint16(0xA409)) // User-defined fault code.
	t.Run("success", func(t *testing.T) {
		conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x72, response},
			propertyExchange{[]uint8{0x82, 0x9F}, 0x72, []Property{
				testProperty(0x82, 0, 0, 'R', 3),
				testProperty(0x9F, 13, 0x80, 0x82, 0x88, 0x89, 0x9F, 0xA0, 0xC3, 0xD0, 0xD1, 0xE0, 0xE1, 0xE3, 0xE8),
			}},
			propertyExchange{[]uint8{0x80, 0x88, 0x89, 0xD0, 0xD1}, 0x72, []Property{
				testProperty(0x80, 0x30),
				testProperty(0x88, 0x42),
				testProperty(0x89, 0xA4, 0x09),
				testProperty(0xD0, 3),
				testProperty(0xD1, 0x45),
			}},
			propertyExchange{[]uint8{0xA0}, 0x72, []Property{testProperty(0xA0, 80)}},
			propertyExchange{[]uint8{0xC3}, 0x72, []Property{testProperty(0xC3, 0x03, 0xE8)}},
			propertyExchange{[]uint8{0xE8}, 0x72, []Property{testProperty(0xE8, 0xFF, 0xFF)}},
			propertyExchange{[]uint8{0xE3}, 0x72, []Property{testProperty(0xE3, 0, 0, 0, 45)}},
		)
		got, err := NewPVPowerGenerationClient(conn).Get(t.Context(), "127.0.0.1", eoj)
		if err != nil || got == nil || !reflect.DeepEqual(got, &want) {
			t.Fatalf("Get = (%+v, %v), want %+v", got, err, want)
		}
	})

	for _, tc := range []struct {
		name     string
		index    int
		property Property
	}{
		{"short power", 0, testProperty(0xE0, 0)},
		{"power overflow", 0, testProperty(0xE0, 0xFF, 0xFE)},
		{"short energy", 1, testProperty(0xE1, 0, 0, 0)},
		{"energy above max", 1, testProperty(0xE1, 0x3B, 0x9A, 0xCA, 0x00)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			props := slices.Clone(response)
			props[tc.index] = tc.property
			conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x72, props})
			if got, err := NewPVPowerGenerationClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
				t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
			}
		})
	}

	t.Run("missing required property", func(t *testing.T) {
		props := slices.Clone(response)
		props[len(props)-1].EDT = nil
		conn := newPropertyTestConnection(t, eoj, propertyExchange{epcs, 0x52, props})
		if got, err := NewPVPowerGenerationClient(conn).Get(t.Context(), "127.0.0.1", eoj); got != nil || err == nil {
			t.Fatalf("Get = (%+v, %v), want an error and no reading", got, err)
		}
	})
}

func TestPVPowerGenerationGetHonorsReleaseSpecificRequirements(t *testing.T) {
	eoj := EOJ{0x02, 0x79, 0x01}
	base := propertyExchange{[]uint8{0xE0, 0xE1}, 0x72, []Property{testProperty(0xE0, 0, 1), testProperty(0xE1, 0, 0, 0, 1)}}
	for _, release := range []byte{'a', 'J', 'K'} {
		t.Run(string(release), func(t *testing.T) {
			exchanges := []propertyExchange{base,
				{[]uint8{0x82, 0x9F}, 0x72, []Property{
					testProperty(0x82, 0, 0, release, 0),
					testProperty(0x9F, 6, 0x80, 0x82, 0x88, 0x9F, 0xE0, 0xE1),
				}},
			}
			if release == 'K' {
				exchanges = append(exchanges,
					propertyExchange{[]uint8{0x80, 0x88, 0xD0, 0xD1}, 0x72, []Property{
						testProperty(0x80, 0x30), testProperty(0x88, 0x42), testProperty(0xD0, 0), testProperty(0xD1, 0x44),
					}},
					propertyExchange{[]uint8{0xE8}, esvGetSNA, []Property{testProperty(0xE8)}},
				)
			} else {
				exchanges = append(exchanges, propertyExchange{[]uint8{0x80, 0x88}, 0x72, []Property{testProperty(0x80, 0x30), testProperty(0x88, 0x42)}})
			}
			conn := newPropertyTestConnection(t, eoj, exchanges...)
			got, err := NewPVPowerGenerationClient(conn).Get(t.Context(), "127.0.0.1", eoj)
			if release != 'K' && (err != nil || got == nil || got.RatedElectricPowerOfgeneration != nil ||
				got.InstantaneousElectricPowerGeneration != 1 || got.CumulativeElectricEnergyOfGeneration != 1) {
				t.Fatalf("Get on release %c = (%+v, %v), want generation readings without optional rated power", release, got, err)
			}
			if release == 'K' && (got != nil || err == nil || !strings.Contains(err.Error(), "EPC 0xE8")) {
				t.Fatalf("Get on release K = (%+v, %v), want missing required rated power", got, err)
			}
		})
	}
}

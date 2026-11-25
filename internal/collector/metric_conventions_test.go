package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type mraProperty struct {
	EPC          string `json:"epc"`
	ShortName    string `json:"shortName"`
	PropertyName struct {
		English string `json:"en"`
	} `json:"propertyName"`
}

func assertMRAMetricConventions(t *testing.T, registry *prometheus.Registry, class uint16, descriptors []metricDescriptor) {
	t.Helper()
	dir, ok := os.LookupEnv("ECHONETLITE_MRA_DIR")
	if !ok {
		t.Skip("set ECHONETLITE_MRA_DIR to the directory printed by third_party/echonet-lite-mra/fetch.sh")
	}
	if dir == "" {
		t.Fatal("ECHONETLITE_MRA_DIR must point to the extracted MRA directory")
	}

	classCode := fmt.Sprintf("0x%04X", class)
	path := filepath.Join(dir, "devices", classCode+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read MRA device class %s: %v", classCode, err)
	}
	var device struct {
		EOJ        string        `json:"eoj"`
		ShortName  string        `json:"shortName"`
		Properties []mraProperty `json:"elProperties"`
	}
	if err := json.Unmarshal(data, &device); err != nil {
		t.Fatalf("decode MRA device class %s: %v", classCode, err)
	}
	if device.EOJ != classCode || device.ShortName == "" {
		t.Fatalf("MRA file %s must describe device class %s and provide its shortName", path, classCode)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("%s: gather metric families: %v", classCode, err)
	}
	unmatched := make(map[string]int, len(families))
	for index, family := range families {
		if name := family.GetName(); name != "echonetlite_collector_success" {
			unmatched[name] = index
		}
	}
	superData, err := os.ReadFile(filepath.Join(dir, "superClass", "0x0000.json"))
	if err != nil {
		t.Fatalf("read MRA superclass: %v", err)
	}
	var superclass struct {
		Properties []mraProperty `json:"elProperties"`
	}
	if err := json.Unmarshal(superData, &superclass); err != nil {
		t.Fatalf("decode MRA superclass: %v", err)
	}
	mraProperties := make(map[string]mraProperty)
	for _, property := range slices.Concat(superclass.Properties, device.Properties) {
		mraProperties[property.ShortName] = property
	}

	className := metricSnakeCase(device.ShortName)
	for _, descriptor := range descriptors {
		mra, ok := mraProperties[descriptor.propertyShortName]
		if !ok {
			t.Errorf("%s: no MRA property with shortName %q", classCode, descriptor.propertyShortName)
			continue
		}
		if mra.PropertyName.English == "" {
			t.Errorf("%s/%s: MRA property %q must provide propertyName.en", classCode, mra.EPC, mra.ShortName)
			continue
		}

		var qualifier string
		wantType := "GAUGE"
		switch descriptor.reading {
		case instantaneousReading:
			qualifier = "instantaneous"
		case cumulativeReading:
			qualifier = "cumulative"
			wantType = "COUNTER"
		}
		words := strings.Split(metricSnakeCase(mra.ShortName), "_")
		if qualifier != "" {
			words = slices.DeleteFunc(words, func(word string) bool { return word == qualifier })
		}
		nameParts := []string{"echonetlite", className, strings.Join(words, "_")}
		if descriptor.unit != "" {
			nameParts = append(nameParts, descriptor.unit)
		}
		if descriptor.reading == cumulativeReading {
			nameParts = append(nameParts, "total")
		}
		wantName := strings.Join(nameParts, "_")
		wantHelp := mra.PropertyName.English + ". EPC=" + mra.EPC
		// Check descriptor metadata even when the fixture has no sample for it.
		wantDesc := fmt.Sprintf("Desc{fqName: %q, help: %q,", wantName, wantHelp)
		if !strings.HasPrefix(descriptor.desc.String(), wantDesc) {
			t.Errorf("%s/%s: got %s, want name %q and HELP %q", classCode, mra.EPC, descriptor.desc, wantName, wantHelp)
		}
		index, ok := unmatched[wantName]
		if !ok {
			// An unsupported or unavailable property has a descriptor and no sample.
			continue
		}
		delete(unmatched, wantName)

		family := families[index]
		if family.GetHelp() != wantHelp {
			t.Errorf("%s/%s: %s HELP = %q, want %q", classCode, mra.EPC, wantName, family.GetHelp(), wantHelp)
		}
		if gotType := family.GetType().String(); gotType != wantType {
			t.Errorf("%s/%s: %s type = %s, want %s (instantaneous readings and states use gauges, cumulative readings use counters)",
				classCode, mra.EPC, wantName, gotType, wantType)
		}
	}
	for _, family := range families {
		name := family.GetName()
		if _, ok := unmatched[name]; ok {
			t.Errorf("%s: unexpected metric family %q; each device metric family must match an expected MRA EPC", classCode, name)
		}
	}
}

var (
	acronymBoundary = regexp.MustCompile(`([A-Z])([A-Z][a-z])`)
	wordBoundary    = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func metricSnakeCase(name string) string {
	name = acronymBoundary.ReplaceAllString(name, "${1}_${2}")
	return strings.ToLower(wordBoundary.ReplaceAllString(name, "${1}_${2}"))
}

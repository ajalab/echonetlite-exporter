package collector

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCollectorsPublishSuccessInSameRegistry(t *testing.T) {
	gasMeter := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x82, 0x01})
	wattHourMeter := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x80, 0x01})
	gasCollector := NewGasMeterCollector(
		func(context.Context, string, echonetlite.EOJ) (*echonetlite.GasMeter, error) {
			return &echonetlite.GasMeter{CumulativeGasConsumption: 100}, nil
		},
		time.Second, []echonetlite.Device{gasMeter},
	)
	wattHourCollector := NewWattHourMeterCollector(
		func(context.Context, string, echonetlite.EOJ) (*echonetlite.WattHourMeter, error) {
			return nil, errors.New("required property unavailable")
		},
		time.Second, []echonetlite.Device{wattHourMeter},
	)
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(gasCollector, wattHourCollector)
	gasCollector.collect(t.Context())
	wattHourCollector.collect(t.Context())

	const want = `
# HELP echonetlite_collector_success Whether the last collection was successful (1 for success, 0 for failure).
# TYPE echonetlite_collector_success gauge
echonetlite_collector_success{eoj="028001",host="192.0.2.10"} 0
echonetlite_collector_success{eoj="028201",host="192.0.2.10"} 1
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(want), "echonetlite_collector_success"); err != nil {
		t.Fatal(err)
	}
}

func TestCollectorRetainsFailedReadAndReplacesSuccessfulRead(t *testing.T) {
	first := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x72, 0x01})
	second := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x72, 0x02})
	type reading struct {
		value *float64
	}
	firstValue, secondValue := new(float64(0)), new(float64(42))
	var firstErr error
	valueDesc := prometheus.NewDesc("test_reading", "Test reading.", deviceLabels, nil)
	c := newCollector(time.Second, []echonetlite.Device{first, second}, []*prometheus.Desc{valueDesc},
		func(_ context.Context, _ string, eoj echonetlite.EOJ) (*reading, error) {
			if eoj == first.EOJ() {
				return &reading{value: firstValue}, firstErr
			}
			return &reading{value: secondValue}, nil
		},
		func(device echonetlite.Device, reading *reading) []prometheus.Metric {
			if reading.value == nil {
				return nil
			}
			return []prometheus.Metric{
				prometheus.MustNewConstMetric(valueDesc, prometheus.GaugeValue, *reading.value, device.Host(), device.EOJ().String()),
			}
		},
	)
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(c)

	const value = "test_reading"
	const status = "echonetlite_collector_success"
	const headers = `
# HELP echonetlite_collector_success Whether the last collection was successful (1 for success, 0 for failure).
# TYPE echonetlite_collector_success gauge
# HELP test_reading Test reading.
# TYPE test_reading gauge
`
	c.collect(t.Context())
	want := headers + `
echonetlite_collector_success{eoj="027201",host="192.0.2.10"} 1
echonetlite_collector_success{eoj="027202",host="192.0.2.10"} 1
test_reading{eoj="027201",host="192.0.2.10"} 0
test_reading{eoj="027202",host="192.0.2.10"} 42
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(want), status, value); err != nil {
		t.Fatal(err)
	}

	firstErr = errors.New("read failed")
	*secondValue = 45
	c.collect(t.Context())
	want = headers + `
echonetlite_collector_success{eoj="027201",host="192.0.2.10"} 0
echonetlite_collector_success{eoj="027202",host="192.0.2.10"} 1
test_reading{eoj="027201",host="192.0.2.10"} 0
test_reading{eoj="027202",host="192.0.2.10"} 45
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(want), status, value); err != nil {
		t.Fatal(err)
	}

	firstValue, firstErr = nil, nil
	c.collect(t.Context())
	want = headers + `
echonetlite_collector_success{eoj="027201",host="192.0.2.10"} 1
echonetlite_collector_success{eoj="027202",host="192.0.2.10"} 1
test_reading{eoj="027202",host="192.0.2.10"} 45
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(want), status, value); err != nil {
		t.Fatal(err)
	}
}

func TestCollectorStartsImmediatelyAndStopsOnCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		first := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x82, 0x01})
		second := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{0x02, 0x82, 0x02})
		calls := make(chan echonetlite.EOJ, 4)
		callCount := 0
		c := newCollector(time.Minute, []echonetlite.Device{first, second}, nil,
			func(ctx context.Context, host string, eoj echonetlite.EOJ) (*float64, error) {
				if host != first.Host() {
					t.Fatalf("Get host = %q, want %q", host, first.Host())
				}
				callCount++
				calls <- eoj
				if callCount > 2 {
					<-ctx.Done()
					return nil, ctx.Err()
				}
				return new(float64(0)), nil
			},
			func(echonetlite.Device, *float64) []prometheus.Metric { return nil },
		)
		c.Start(ctx)
		synctest.Wait()
		if len(calls) != 2 {
			t.Fatalf("initial poll got %d calls, want 2", len(calls))
		}
		if gotFirst, gotSecond := <-calls, <-calls; gotFirst != first.EOJ() || gotSecond != second.EOJ() {
			t.Fatalf("initial poll calls = %v, %v", gotFirst, gotSecond)
		}

		time.Sleep(time.Minute)
		synctest.Wait()
		if len(calls) != 1 {
			t.Fatalf("next poll got %d calls, want one", len(calls))
		}
		if got := <-calls; got != first.EOJ() {
			t.Fatalf("next poll call = %v, want %v", got, first.EOJ())
		}

		cancel()
		synctest.Wait()
		time.Sleep(2 * time.Minute)
		if len(calls) != 0 {
			t.Fatal("Get was called after cancellation")
		}
	})
}

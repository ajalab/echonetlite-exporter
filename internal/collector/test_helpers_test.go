package collector

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type testCollector interface {
	prometheus.Collector
	collect(context.Context)
	metricDescriptors() []metricDescriptor
}

func (c *collector) metricDescriptors() []metricDescriptor {
	return c.descs
}

func assertPublishedMetrics[T any, C testCollector](
	t *testing.T,
	build func(func(context.Context, string, echonetlite.EOJ) (*T, error), time.Duration, []echonetlite.Device) C,
	class uint16,
	reading *T,
	want string,
) {
	t.Helper()
	device := echonetlite.NewDevice("192.0.2.10", echonetlite.EOJ{byte(class >> 8), byte(class), 1})
	c := build(func(context.Context, string, echonetlite.EOJ) (*T, error) {
		return reading, nil
	}, time.Second, []echonetlite.Device{device})
	c.collect(t.Context())

	wantSuccess := fmt.Sprintf(`
# HELP echonetlite_collector_success Whether the last collection was successful (1 for success, 0 for failure).
# TYPE echonetlite_collector_success gauge
echonetlite_collector_success{eoj=%q,host=%q} 1
`, device.EOJ().String(), device.Host())
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(c)
	t.Run("MRAConventions", func(t *testing.T) {
		assertMRAMetricConventions(t, registry, class, c.metricDescriptors())
	})
	if err := testutil.GatherAndCompare(registry, strings.NewReader(wantSuccess+want)); err != nil {
		t.Fatal(err)
	}
}

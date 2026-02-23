package collector

import (
	"math"
	"testing"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func assertCollectSuccess(t *testing.T, collectMetrics *CollectMetrics, device echonetlite.Device, want float64) {
	t.Helper()
	got := testutil.ToFloat64(
		collectMetrics.successGauge.WithLabelValues(device.Host(), device.EOJ().String()),
	)
	assertFloatEqual(t, got, want)
}

func assertFloatEqual(t *testing.T, got, want float64) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func assertFloatApprox(t *testing.T, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("expected about %v, got %v (tol=%v)", want, got, tol)
	}
}

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CollectMetrics struct {
	successGauge *prometheus.GaugeVec
}

func NewCollectMetrics() *CollectMetrics {
	return &CollectMetrics{
		successGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_collector_success",
				Help: "Whether the last collection was successful (1 for success, 0 for failure).",
			},
			[]string{"host", "eoj"},
		),
	}
}

func (m *CollectMetrics) Collector() prometheus.Collector {
	return m.successGauge
}

func (m *CollectMetrics) SetSuccess(host, eoj string, success bool) {
	if success {
		m.successGauge.WithLabelValues(host, eoj).Set(1)
	} else {
		m.successGauge.WithLabelValues(host, eoj).Set(0)
	}
}

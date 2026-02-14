package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

type collectMetrics struct {
	collector    string
	successGauge *prometheus.GaugeVec
}

func newCollectMetrics(collector string) *collectMetrics {
	return &collectMetrics{
		collector: collector,
		successGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_collector_success",
				Help: "Whether the last collection was successful (1 for success, 0 for failure).",
			},
			[]string{"collector", "host", "eoj"},
		),
	}
}

func (m *collectMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.successGauge.Describe(ch)
}

func (m *collectMetrics) Collect(ch chan<- prometheus.Metric) {
	m.successGauge.Collect(ch)
}

func (m *collectMetrics) SetSuccess(host, eoj string, success bool) {
	if success {
		m.successGauge.WithLabelValues(m.collector, host, eoj).Set(1)
	} else {
		m.successGauge.WithLabelValues(m.collector, host, eoj).Set(0)
	}
}

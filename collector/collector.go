package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

type collectMetrics struct {
	successGauge *prometheus.GaugeVec
}

func NewPollMetrics() *collectMetrics {
	return &collectMetrics{
		successGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_collect_success",
				Help: "Whether the last collection was successful (1 for success, 0 for failure).",
			},
			[]string{"host", "eoj"},
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
		m.successGauge.WithLabelValues(host, eoj).Set(1)
	} else {
		m.successGauge.WithLabelValues(host, eoj).Set(0)
	}
}

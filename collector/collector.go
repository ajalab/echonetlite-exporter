package collector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	requestTimeout = 10 * time.Second
)

type PollMetrics struct {
	pollErrorsTotal          *prometheus.CounterVec
	lastPollTimestampSeconds *prometheus.GaugeVec
}

func NewPollMetrics() *PollMetrics {
	return &PollMetrics{
		pollErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "echonetlite_poll_errors_total",
				Help: "Total number of polling errors.",
			},
			[]string{"host", "eoj"},
		),
		lastPollTimestampSeconds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_poll_last_successful_time_seconds",
				Help: "Unix timestamp of the last successful poll.",
			},
			[]string{"host", "eoj"},
		),
	}
}

func (m *PollMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.pollErrorsTotal.Describe(ch)
	m.lastPollTimestampSeconds.Describe(ch)
}

func (m *PollMetrics) Collect(ch chan<- prometheus.Metric) {
	m.pollErrorsTotal.Collect(ch)
	m.lastPollTimestampSeconds.Collect(ch)
}

func (m *PollMetrics) IncPollError(host, eoj string) {
	m.pollErrorsTotal.WithLabelValues(host, eoj).Inc()
}

func (m *PollMetrics) SetLastPollTimestamp(host, eoj string, ts time.Time) {
	m.lastPollTimestampSeconds.WithLabelValues(host, eoj).Set(float64(ts.Unix()))
}

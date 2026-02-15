package collector

import (
	"context"
	"log/slog"
	"time"

	"echonetlite-exporter/echonetlite"

	"github.com/prometheus/client_golang/prometheus"
)

type PVPowerGenerationCollector struct {
	objectUpdates <-chan []*echonetlite.PVPowerGeneration
	interval      time.Duration
	timeout       time.Duration

	instantaneousPowerGauge *prometheus.GaugeVec
	cumulativeEnergyGauge   *prometheus.GaugeVec
	collectMetrics          *CollectMetrics
}

func NewPVPowerGenerationCollector(
	conn *echonetlite.Connection,
	updates <-chan []*echonetlite.PVPowerGeneration,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
) *PVPowerGenerationCollector {
	return &PVPowerGenerationCollector{
		objectUpdates: updates,
		interval:      interval,
		timeout:       timeout,
		instantaneousPowerGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_pv_power_generation_electric_power_generation_watts",
				Help: "Instantaneous generated power.",
			},
			[]string{"host", "eoj"},
		),
		cumulativeEnergyGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_pv_power_generation_electric_energy_generation_joules_total",
				Help: "Cumulative generated energy.",
			},
			[]string{"host", "eoj"},
		),
		collectMetrics: collectMetrics,
	}
}

func (c *PVPowerGenerationCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx)
}

func (c *PVPowerGenerationCollector) Describe(ch chan<- *prometheus.Desc) {
	c.instantaneousPowerGauge.Describe(ch)
	c.cumulativeEnergyGauge.Describe(ch)
}

func (c *PVPowerGenerationCollector) Collect(ch chan<- prometheus.Metric) {
	c.instantaneousPowerGauge.Collect(ch)
	c.cumulativeEnergyGauge.Collect(ch)
}

func (c *PVPowerGenerationCollector) collectLoop(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	var pvs []*echonetlite.PVPowerGeneration
	for {
		select {
		case <-ticker.C:
			c.collect(ctx, pvs)
		case updated := <-c.objectUpdates:
			pvs = updated
			c.collect(ctx, pvs)
		case <-ctx.Done():
			return
		}
	}
}

func (c *PVPowerGenerationCollector) collect(ctx context.Context, pvs []*echonetlite.PVPowerGeneration) {
	for _, pv := range pvs {
		reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
		props, err := pv.Get(reqCtx)
		cancel()

		if err != nil {
			c.collectMetrics.SetSuccess(pv.Host(), pv.EOJ().String(), false)
			slog.Warn("failed to collect stats", "host", pv.Host(), "eoj", pv.EOJ().String(), "err", err)
			continue
		}
		c.collectMetrics.SetSuccess(pv.Host(), pv.EOJ().String(), true)

		c.updateMetrics(pv, props)
	}
}

func (c *PVPowerGenerationCollector) updateMetrics(
	pv *echonetlite.PVPowerGeneration,
	props *echonetlite.PVPowerGenerationProps,
) {
	c.instantaneousPowerGauge.WithLabelValues(pv.Host(), pv.EOJ().String()).Set(float64(props.InstantaneousElectricPowerGeneration))
	kWh := float64(props.CumulativeElectricEnergyOfGeneration) * 0.001
	joules := kWh * 3600000
	c.cumulativeEnergyGauge.WithLabelValues(pv.Host(), pv.EOJ().String()).Set(joules)
}

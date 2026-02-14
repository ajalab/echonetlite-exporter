package collector

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"echonetlite-exporter/echonetlite"

	"github.com/prometheus/client_golang/prometheus"
)

type PowerDistributionBoardMeterCollector struct {
	objectUpdates <-chan []*echonetlite.PowerDistributionBoardMetering
	interval      time.Duration
	timeout       time.Duration

	instantaneousElectricPowerSimplexGauge *prometheus.GaugeVec
	cumulativeElectricEnergySimplexGauge   *prometheus.GaugeVec
	collectMetrics                         *collectMetrics
}

func NewPowerDistributionBoardMeterCollector(
	conn *echonetlite.Connection,
	updates <-chan []*echonetlite.PowerDistributionBoardMetering,
	interval time.Duration,
	timeout time.Duration,
) *PowerDistributionBoardMeterCollector {
	return &PowerDistributionBoardMeterCollector{
		objectUpdates: updates,
		interval:      interval,
		timeout:       timeout,
		instantaneousElectricPowerSimplexGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_pdbm_electric_power_simplex_watts",
				Help: "Instantaneous electric power per channel (simplex).",
			},
			[]string{"host", "eoj", "channel"},
		),
		cumulativeElectricEnergySimplexGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_pdbm_electric_energy_simplex_joules_total",
				Help: "Cumulative amount of electric power consumption (simplex)",
			},
			[]string{"host", "eoj", "channel"},
		),
		collectMetrics: newCollectMetrics("pdbm"),
	}
}

func (c *PowerDistributionBoardMeterCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx)
}

func (c *PowerDistributionBoardMeterCollector) Describe(ch chan<- *prometheus.Desc) {
	c.cumulativeElectricEnergySimplexGauge.Describe(ch)
	c.instantaneousElectricPowerSimplexGauge.Describe(ch)
	c.collectMetrics.Describe(ch)
}

func (c *PowerDistributionBoardMeterCollector) Collect(ch chan<- prometheus.Metric) {
	c.cumulativeElectricEnergySimplexGauge.Collect(ch)
	c.instantaneousElectricPowerSimplexGauge.Collect(ch)
	c.collectMetrics.Collect(ch)
}

func (c *PowerDistributionBoardMeterCollector) collectLoop(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	var pdbms []*echonetlite.PowerDistributionBoardMetering
	for {
		select {
		case <-ticker.C:
			c.collect(ctx, pdbms)
		case updated := <-c.objectUpdates:
			pdbms = updated
			c.collect(ctx, pdbms)
		case <-ctx.Done():
			return
		}
	}
}

func (c *PowerDistributionBoardMeterCollector) collect(ctx context.Context, pdbms []*echonetlite.PowerDistributionBoardMetering) {
	for _, pdbm := range pdbms {
		reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
		props, err := pdbm.Get(reqCtx)
		cancel()

		if err != nil {
			c.collectMetrics.SetSuccess(pdbm.Host(), pdbm.EOJ().String(), false)
			slog.Warn("failed to collect stats", "host", pdbm.Host(), "eoj", pdbm.EOJ().String(), "err", err)
			continue
		}
		c.collectMetrics.SetSuccess(pdbm.Host(), pdbm.EOJ().String(), true)

		c.updateMetrics(pdbm, props)
	}
}

func (c *PowerDistributionBoardMeterCollector) updateMetrics(
	pdbm *echonetlite.PowerDistributionBoardMetering,
	props *echonetlite.PowerDistributionBoardMeteringProps,
) {
	{
		start := int(props.InstantaneousElectricPowerListSimplex.StartChannel)
		for i, val := range props.InstantaneousElectricPowerListSimplex.InstantaneousElectricPower {
			channel := fmt.Sprintf("%d", start+i)
			c.instantaneousElectricPowerSimplexGauge.WithLabelValues(pdbm.Host(), pdbm.EOJ().String(), channel).Set(float64(val))
		}
	}
	{
		start := int(props.CumulativeElectricEnergyListSimplex.StartChannel)
		for i, val := range props.CumulativeElectricEnergyListSimplex.ElectricEnergy {
			channel := fmt.Sprintf("%d", start+i)
			cumulativeValue := float64(val) * float64(props.UnitForCumulativeEnergy) * 1000 * 3600 // kWh to J
			c.cumulativeElectricEnergySimplexGauge.WithLabelValues(pdbm.Host(), pdbm.EOJ().String(), channel).Set(cumulativeValue)
		}
	}
}

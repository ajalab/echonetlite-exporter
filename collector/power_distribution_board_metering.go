package collector

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"echonetlite-exporter/echonetlite"

	"github.com/prometheus/client_golang/prometheus"
)

type PowerDistributionBoardMeteringCollector struct {
	objectUpdates <-chan []*echonetlite.PowerDistributionBoardMetering
	interval      time.Duration
	timeout       time.Duration

	instantaneousElectricPowerSimplexGauge *prometheus.GaugeVec
	cumulativeElectricEnergySimplexDesc    *prometheus.Desc
	cumulativeElectricEnergySimplex        map[cumulativeElectricEnergyKey]float64
	cumulativeElectricEnergySimplexMu      sync.Mutex
	collectMetrics                         *CollectMetrics
}

type cumulativeElectricEnergyKey struct {
	host    string
	eoj     string
	channel string
}

func NewPowerDistributionBoardMeteringCollector(
	conn *echonetlite.Connection,
	updates <-chan []*echonetlite.PowerDistributionBoardMetering,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
) *PowerDistributionBoardMeteringCollector {
	return &PowerDistributionBoardMeteringCollector{
		objectUpdates: updates,
		interval:      interval,
		timeout:       timeout,
		instantaneousElectricPowerSimplexGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_power_distribution_board_metering_electric_power_simplex_watts",
				Help: "Instantaneous electric power per channel (simplex).",
			},
			[]string{"host", "eoj", "channel"},
		),
		cumulativeElectricEnergySimplexDesc: prometheus.NewDesc(
			"echonetlite_power_distribution_board_metering_electric_energy_simplex_joules_total",
			"Cumulative amount of electric power consumption (simplex)",
			[]string{"host", "eoj", "channel"},
			nil,
		),
		cumulativeElectricEnergySimplex: make(map[cumulativeElectricEnergyKey]float64),
		collectMetrics: collectMetrics,
	}
}

func (c *PowerDistributionBoardMeteringCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx)
}

func (c *PowerDistributionBoardMeteringCollector) Describe(ch chan<- *prometheus.Desc) {
	c.instantaneousElectricPowerSimplexGauge.Describe(ch)
	ch <- c.cumulativeElectricEnergySimplexDesc
}

func (c *PowerDistributionBoardMeteringCollector) Collect(ch chan<- prometheus.Metric) {
	c.instantaneousElectricPowerSimplexGauge.Collect(ch)
	c.cumulativeElectricEnergySimplexMu.Lock()
	for key, value := range c.cumulativeElectricEnergySimplex {
		ch <- prometheus.MustNewConstMetric(
			c.cumulativeElectricEnergySimplexDesc,
			prometheus.CounterValue,
			value,
			key.host,
			key.eoj,
			key.channel,
		)
	}
	c.cumulativeElectricEnergySimplexMu.Unlock()
}

func (c *PowerDistributionBoardMeteringCollector) collectLoop(ctx context.Context) {
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

func (c *PowerDistributionBoardMeteringCollector) collect(ctx context.Context, pdbms []*echonetlite.PowerDistributionBoardMetering) {
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

func (c *PowerDistributionBoardMeteringCollector) updateMetrics(
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
		c.cumulativeElectricEnergySimplexMu.Lock()
		for i, val := range props.CumulativeElectricEnergyListSimplex.ElectricEnergy {
			channel := fmt.Sprintf("%d", start+i)
			cumulativeValue := float64(val) * float64(props.UnitForCumulativeEnergy) * 1000 * 3600 // kWh to J
			key := cumulativeElectricEnergyKey{
				host:    pdbm.Host(),
				eoj:     pdbm.EOJ().String(),
				channel: channel,
			}
			c.cumulativeElectricEnergySimplex[key] = cumulativeValue
		}
		c.cumulativeElectricEnergySimplexMu.Unlock()
	}
}

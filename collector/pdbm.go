package collector

import (
	"context"
	"fmt"
	"log"
	"time"

	"echonetlite-exporter/echonetlite"

	"github.com/prometheus/client_golang/prometheus"
)

const pollInterval = 15 * time.Second

type PowerDistributionBoardMeterCollector struct {
	objectUpdates <-chan []*echonetlite.PowerDistributionBoardMetering

	instantaneousElectricPowerSimplexGauge *prometheus.GaugeVec
	cumulativeElectricEnergySimplexGauge   *prometheus.GaugeVec
	pollMetrics                            *PollMetrics
}

func NewPowerDistributionBoardMeterCollector(conn *echonetlite.Connection, updates <-chan []*echonetlite.PowerDistributionBoardMetering) *PowerDistributionBoardMeterCollector {
	return &PowerDistributionBoardMeterCollector{
		objectUpdates: updates,
		instantaneousElectricPowerSimplexGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_pdbm_instantaneous_electric_power_simplex_watts",
				Help: "Instantaneous electric power per channel (simplex).",
			},
			[]string{"host", "eoj", "channel"},
		),
		cumulativeElectricEnergySimplexGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_pdbm_cumulative_electric_energy_simplex_watts",
				Help: "Cumulative amount of electric power consumption (simplex)",
			},
			[]string{"host", "eoj", "channel"},
		),
		pollMetrics: NewPollMetrics(),
	}
}

func (c *PowerDistributionBoardMeterCollector) Start(ctx context.Context) {
	if c.objectUpdates == nil {
		log.Printf("pdbm collector has no object updates channel; polling disabled")
		return
	}
	go c.pollLoop(ctx)
}

func (c *PowerDistributionBoardMeterCollector) Describe(ch chan<- *prometheus.Desc) {
	c.cumulativeElectricEnergySimplexGauge.Describe(ch)
	c.instantaneousElectricPowerSimplexGauge.Describe(ch)
	c.pollMetrics.Describe(ch)
}

func (c *PowerDistributionBoardMeterCollector) Collect(ch chan<- prometheus.Metric) {
	c.cumulativeElectricEnergySimplexGauge.Collect(ch)
	c.instantaneousElectricPowerSimplexGauge.Collect(ch)
	c.pollMetrics.Collect(ch)
}

func (c *PowerDistributionBoardMeterCollector) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var pdbms []*echonetlite.PowerDistributionBoardMetering
	for {
		select {
		case <-ticker.C:
			c.runPoll(ctx, pdbms)
		case updated := <-c.objectUpdates:
			pdbms = updated
			c.runPoll(ctx, pdbms)
		case <-ctx.Done():
			return
		}
	}
}

func (c *PowerDistributionBoardMeterCollector) runPoll(ctx context.Context, pdbms []*echonetlite.PowerDistributionBoardMetering) {
	for _, pdbm := range pdbms {
		reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		props, err := pdbm.Get(reqCtx)
		cancel()

		if err != nil {
			c.pollMetrics.IncPollError(pdbm.Host(), pdbm.EOJ().String())
			log.Printf("poll error for %s: %v", pdbm.Host(), err)
			continue
		}

		c.updateMetrics(pdbm, props)
		c.pollMetrics.SetLastPollTimestamp(pdbm.Host(), pdbm.EOJ().String(), time.Now())
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
			cumulativeValue := float64(val) * float64(props.UnitForCumulativeEnergy) * 1000
			c.cumulativeElectricEnergySimplexGauge.WithLabelValues(pdbm.Host(), pdbm.EOJ().String(), channel).Set(cumulativeValue)
		}
	}
}

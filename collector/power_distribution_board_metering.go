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
	interval time.Duration
	timeout  time.Duration
	client   *echonetlite.PowerDistributionBoardMeteringClient
	targets  []echonetlite.Device

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
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *PowerDistributionBoardMeteringCollector {
	return &PowerDistributionBoardMeteringCollector{
		interval: interval,
		timeout:  timeout,
		client:   echonetlite.NewPowerDistributionBoardMeteringClient(conn),
		targets:  targets,
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
		collectMetrics:                  collectMetrics,
	}
}

func (c *PowerDistributionBoardMeteringCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
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

func (c *PowerDistributionBoardMeteringCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	c.collect(ctx, targets)
	for {
		select {
		case <-ticker.C:
			c.collect(ctx, targets)
		case <-ctx.Done():
			return
		}
	}
}

func (c *PowerDistributionBoardMeteringCollector) collect(ctx context.Context, devices []echonetlite.Device) {
	for _, device := range devices {
		reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
		props, err := c.client.Get(reqCtx, device.Host(), device.EOJ())
		cancel()

		if err != nil {
			c.collectMetrics.SetSuccess(device.Host(), device.EOJ().String(), false)
			slog.Warn("failed to collect stats", "host", device.Host(), "eoj", device.EOJ().String(), "err", err)
			continue
		}
		c.collectMetrics.SetSuccess(device.Host(), device.EOJ().String(), true)

		c.updateMetrics(device, props)
	}
}

func (c *PowerDistributionBoardMeteringCollector) updateMetrics(
	device echonetlite.Device,
	pdbm *echonetlite.PowerDistributionBoardMetering,
) {
	{
		start := int(pdbm.InstantaneousElectricPowerListSimplex.StartChannel)
		for i, val := range pdbm.InstantaneousElectricPowerListSimplex.InstantaneousElectricPower {
			channel := fmt.Sprintf("%d", start+i)
			c.instantaneousElectricPowerSimplexGauge.WithLabelValues(device.Host(), device.EOJ().String(), channel).Set(float64(val))
		}
	}
	{
		start := int(pdbm.CumulativeElectricEnergyListSimplex.StartChannel)
		c.cumulativeElectricEnergySimplexMu.Lock()
		for i, val := range pdbm.CumulativeElectricEnergyListSimplex.ElectricEnergy {
			channel := fmt.Sprintf("%d", start+i)
			cumulativeValue := float64(val) * float64(pdbm.UnitForCumulativeEnergy) * 1000 * 3600 // kWh to J
			key := cumulativeElectricEnergyKey{
				host:    device.Host(),
				eoj:     device.EOJ().String(),
				channel: channel,
			}
			c.cumulativeElectricEnergySimplex[key] = cumulativeValue
		}
		c.cumulativeElectricEnergySimplexMu.Unlock()
	}
}

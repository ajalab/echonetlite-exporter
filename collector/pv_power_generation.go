package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"echonetlite-exporter/echonetlite"

	"github.com/prometheus/client_golang/prometheus"
)

type PVPowerGenerationCollector struct {
	interval time.Duration
	timeout  time.Duration
	client   *echonetlite.PVPowerGenerationClient
	targets  []echonetlite.Device

	instantaneousPowerGauge *prometheus.GaugeVec
	cumulativeEnergyDesc    *prometheus.Desc
	cumulativeEnergy        map[cumulativeEnergyKey]float64
	cumulativeEnergyMu      sync.Mutex
	collectMetrics          *CollectMetrics
}

type cumulativeEnergyKey struct {
	host string
	eoj  string
}

func NewPVPowerGenerationCollector(
	conn *echonetlite.Connection,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *PVPowerGenerationCollector {
	return &PVPowerGenerationCollector{
		interval: interval,
		timeout:  timeout,
		client:   echonetlite.NewPVPowerGenerationClient(conn),
		targets:  targets,
		instantaneousPowerGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_pv_power_generation_electric_power_generation_watts",
				Help: "Instantaneous generated power.",
			},
			[]string{"host", "eoj"},
		),
		cumulativeEnergyDesc: prometheus.NewDesc(
			"echonetlite_pv_power_generation_electric_energy_generation_joules_total",
			"Cumulative generated energy.",
			[]string{"host", "eoj"},
			nil,
		),
		cumulativeEnergy: make(map[cumulativeEnergyKey]float64),
		collectMetrics:   collectMetrics,
	}
}

func (c *PVPowerGenerationCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *PVPowerGenerationCollector) Describe(ch chan<- *prometheus.Desc) {
	c.instantaneousPowerGauge.Describe(ch)
	ch <- c.cumulativeEnergyDesc
}

func (c *PVPowerGenerationCollector) Collect(ch chan<- prometheus.Metric) {
	c.instantaneousPowerGauge.Collect(ch)
	c.cumulativeEnergyMu.Lock()
	for key, value := range c.cumulativeEnergy {
		ch <- prometheus.MustNewConstMetric(
			c.cumulativeEnergyDesc,
			prometheus.CounterValue,
			value,
			key.host,
			key.eoj,
		)
	}
	c.cumulativeEnergyMu.Unlock()
}

func (c *PVPowerGenerationCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
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

func (c *PVPowerGenerationCollector) collect(ctx context.Context, devices []echonetlite.Device) {
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

func (c *PVPowerGenerationCollector) updateMetrics(
	device echonetlite.Device,
	pvpg *echonetlite.PVPowerGeneration,
) {
	c.instantaneousPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()).Set(float64(pvpg.InstantaneousElectricPowerGeneration))

	kWh := float64(pvpg.CumulativeElectricEnergyOfGeneration) * 0.001
	joules := kWh * 3600000
	key := cumulativeEnergyKey{
		host: device.Host(),
		eoj:  device.EOJ().String(),
	}
	c.cumulativeEnergyMu.Lock()
	c.cumulativeEnergy[key] = joules
	c.cumulativeEnergyMu.Unlock()
}

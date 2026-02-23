package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type MultipleInputPCSCollector struct {
	interval time.Duration
	timeout  time.Duration
	get      func(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.MultipleInputPCS, error)
	targets  []echonetlite.Device

	instantaneousElectricPowerGauge *prometheus.GaugeVec
	normalDirectionEnergyDesc       *prometheus.Desc
	normalDirectionEnergy           map[cumulativeMultipleInputPCSEnergyKey]float64
	reverseDirectionEnergyDesc      *prometheus.Desc
	reverseDirectionEnergy          map[cumulativeMultipleInputPCSEnergyKey]float64
	metricsStateMu                  sync.Mutex
	collectMetrics                  *CollectMetrics
}

type cumulativeMultipleInputPCSEnergyKey struct {
	host string
	eoj  string
}

func NewMultipleInputPCSCollector(
	conn *echonetlite.Connection,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *MultipleInputPCSCollector {
	client := echonetlite.NewMultipleInputPCSClient(conn)
	return &MultipleInputPCSCollector{
		interval: interval,
		timeout:  timeout,
		get:      client.Get,
		targets:  targets,
		instantaneousElectricPowerGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_multiple_input_pcs_electric_power_watts",
				Help: "Measured instantaneous electric power.",
			},
			[]string{"host", "eoj"},
		),
		normalDirectionEnergyDesc: prometheus.NewDesc(
			"echonetlite_multiple_input_pcs_normal_direction_electric_energy_joules_total",
			"Cumulative electric energy in the normal direction.",
			[]string{"host", "eoj"},
			nil,
		),
		normalDirectionEnergy: make(map[cumulativeMultipleInputPCSEnergyKey]float64),
		reverseDirectionEnergyDesc: prometheus.NewDesc(
			"echonetlite_multiple_input_pcs_reverse_direction_electric_energy_joules_total",
			"Cumulative electric energy in the reverse direction.",
			[]string{"host", "eoj"},
			nil,
		),
		reverseDirectionEnergy: make(map[cumulativeMultipleInputPCSEnergyKey]float64),
		collectMetrics:         collectMetrics,
	}
}

func (c *MultipleInputPCSCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *MultipleInputPCSCollector) Describe(ch chan<- *prometheus.Desc) {
	c.instantaneousElectricPowerGauge.Describe(ch)
	ch <- c.normalDirectionEnergyDesc
	ch <- c.reverseDirectionEnergyDesc
}

func (c *MultipleInputPCSCollector) Collect(ch chan<- prometheus.Metric) {
	c.instantaneousElectricPowerGauge.Collect(ch)
	c.metricsStateMu.Lock()
	for key, value := range c.normalDirectionEnergy {
		ch <- prometheus.MustNewConstMetric(
			c.normalDirectionEnergyDesc,
			prometheus.CounterValue,
			value,
			key.host,
			key.eoj,
		)
	}
	for key, value := range c.reverseDirectionEnergy {
		ch <- prometheus.MustNewConstMetric(
			c.reverseDirectionEnergyDesc,
			prometheus.CounterValue,
			value,
			key.host,
			key.eoj,
		)
	}
	c.metricsStateMu.Unlock()
}

func (c *MultipleInputPCSCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
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

func (c *MultipleInputPCSCollector) collect(ctx context.Context, devices []echonetlite.Device) {
	for _, device := range devices {
		reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
		props, err := c.get(reqCtx, device.Host(), device.EOJ())
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

func (c *MultipleInputPCSCollector) updateMetrics(
	device echonetlite.Device,
	mipcs *echonetlite.MultipleInputPCS,
) {
	c.instantaneousElectricPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()).Set(
		float64(mipcs.InstantaneousElectricPower),
	)

	key := cumulativeMultipleInputPCSEnergyKey{
		host: device.Host(),
		eoj:  device.EOJ().String(),
	}
	c.metricsStateMu.Lock()
	c.normalDirectionEnergy[key] = float64(mipcs.NormalDirectionElectricEnergy) * 3600.0
	c.reverseDirectionEnergy[key] = float64(mipcs.ReverseDirectionElectricEnergy) * 3600.0
	c.metricsStateMu.Unlock()
}

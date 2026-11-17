package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type GasMeterCollector struct {
	interval time.Duration
	timeout  time.Duration
	client   gasMeterClient
	targets  []echonetlite.Device

	consumptionDesc *prometheus.Desc
	consumption     map[echonetlite.Device]float64
	consumptionMu   sync.Mutex
	collectMetrics  *CollectMetrics
}

type gasMeterClient interface {
	Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.GasMeter, error)
}

func NewGasMeterCollector(
	client gasMeterClient,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *GasMeterCollector {
	return &GasMeterCollector{
		interval: interval,
		timeout:  timeout,
		client:   client,
		targets:  targets,
		consumptionDesc: prometheus.NewDesc(
			"echonetlite_gas_meter_gas_consumption_cubic_meters_total",
			"Cumulative amount of gas consumption measurement value.",
			[]string{"host", "eoj"},
			nil,
		),
		consumption:    make(map[echonetlite.Device]float64),
		collectMetrics: collectMetrics,
	}
}

func (c *GasMeterCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *GasMeterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.consumptionDesc
}

func (c *GasMeterCollector) Collect(ch chan<- prometheus.Metric) {
	c.consumptionMu.Lock()
	defer c.consumptionMu.Unlock()
	for device, value := range c.consumption {
		ch <- prometheus.MustNewConstMetric(
			c.consumptionDesc,
			prometheus.CounterValue,
			value,
			device.Host(),
			device.EOJ().String(),
		)
	}
}

func (c *GasMeterCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
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

func (c *GasMeterCollector) collect(ctx context.Context, devices []echonetlite.Device) {
	for _, device := range devices {
		if ctx.Err() != nil {
			return
		}
		reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
		reading, err := c.client.Get(reqCtx, device.Host(), device.EOJ())
		cancel()

		if err != nil {
			c.collectMetrics.SetSuccess(device.Host(), device.EOJ().String(), false)
			slog.Warn("failed to collect stats", "host", device.Host(), "eoj", device.EOJ().String(), "err", err)
			continue
		}
		c.collectMetrics.SetSuccess(device.Host(), device.EOJ().String(), true)
		c.updateMetrics(device, reading)
	}
}

func (c *GasMeterCollector) updateMetrics(device echonetlite.Device, reading *echonetlite.GasMeter) {
	cubicMeters := float64(reading.CumulativeGasConsumption) / 1_000
	c.consumptionMu.Lock()
	c.consumption[device] = cubicMeters
	c.consumptionMu.Unlock()
}

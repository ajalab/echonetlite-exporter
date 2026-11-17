package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type WattHourMeterCollector struct {
	interval time.Duration
	timeout  time.Duration
	client   wattHourMeterClient
	targets  []echonetlite.Device

	energyDesc     *prometheus.Desc
	energy         map[echonetlite.Device]float64
	energyMu       sync.Mutex
	collectMetrics *CollectMetrics
}

type wattHourMeterClient interface {
	Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.WattHourMeter, error)
}

func NewWattHourMeterCollector(
	client wattHourMeterClient,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *WattHourMeterCollector {
	return &WattHourMeterCollector{
		interval: interval,
		timeout:  timeout,
		client:   client,
		targets:  targets,
		energyDesc: prometheus.NewDesc(
			"echonetlite_watt_hour_meter_electric_energy_joules_total",
			"Cumulative amounts of electric energy measurement value.",
			[]string{"host", "eoj"},
			nil,
		),
		energy:         make(map[echonetlite.Device]float64),
		collectMetrics: collectMetrics,
	}
}

func (c *WattHourMeterCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *WattHourMeterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.energyDesc
}

func (c *WattHourMeterCollector) Collect(ch chan<- prometheus.Metric) {
	c.energyMu.Lock()
	defer c.energyMu.Unlock()
	for device, value := range c.energy {
		ch <- prometheus.MustNewConstMetric(
			c.energyDesc,
			prometheus.CounterValue,
			value,
			device.Host(),
			device.EOJ().String(),
		)
	}
}

func (c *WattHourMeterCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
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

func (c *WattHourMeterCollector) collect(ctx context.Context, devices []echonetlite.Device) {
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

func (c *WattHourMeterCollector) updateMetrics(device echonetlite.Device, reading *echonetlite.WattHourMeter) {
	joules := float64(reading.CumulativeElectricEnergy) * reading.CumulativeAmountsOfElectricEnergyUnit * 3_600_000
	c.energyMu.Lock()
	c.energy[device] = joules
	c.energyMu.Unlock()
}

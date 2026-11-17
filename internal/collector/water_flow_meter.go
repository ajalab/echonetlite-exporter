package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type WaterFlowMeterCollector struct {
	interval time.Duration
	timeout  time.Duration
	client   waterFlowMeterClient
	targets  []echonetlite.Device

	volumeDesc     *prometheus.Desc
	volume         map[echonetlite.Device]float64
	volumeMu       sync.Mutex
	collectMetrics *CollectMetrics
}

type waterFlowMeterClient interface {
	Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.WaterFlowMeter, error)
}

func NewWaterFlowMeterCollector(
	client waterFlowMeterClient,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *WaterFlowMeterCollector {
	return &WaterFlowMeterCollector{
		interval: interval,
		timeout:  timeout,
		client:   client,
		targets:  targets,
		volumeDesc: prometheus.NewDesc(
			"echonetlite_water_flow_meter_flowing_water_cubic_meters_total",
			"Measured cumulative amount of flowing water.",
			[]string{"host", "eoj"},
			nil,
		),
		volume:         make(map[echonetlite.Device]float64),
		collectMetrics: collectMetrics,
	}
}

func (c *WaterFlowMeterCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *WaterFlowMeterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.volumeDesc
}

func (c *WaterFlowMeterCollector) Collect(ch chan<- prometheus.Metric) {
	c.volumeMu.Lock()
	defer c.volumeMu.Unlock()
	for device, value := range c.volume {
		ch <- prometheus.MustNewConstMetric(
			c.volumeDesc,
			prometheus.CounterValue,
			value,
			device.Host(),
			device.EOJ().String(),
		)
	}
}

func (c *WaterFlowMeterCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
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

func (c *WaterFlowMeterCollector) collect(ctx context.Context, devices []echonetlite.Device) {
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

func (c *WaterFlowMeterCollector) updateMetrics(device echonetlite.Device, reading *echonetlite.WaterFlowMeter) {
	cubicMeters := float64(reading.CumulativeFlowingWater) * reading.UnitForCumulativeFlowingWater
	c.volumeMu.Lock()
	c.volume[device] = cubicMeters
	c.volumeMu.Unlock()
}

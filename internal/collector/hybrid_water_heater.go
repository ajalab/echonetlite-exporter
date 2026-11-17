package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type HybridWaterHeaterCollector struct {
	interval time.Duration
	timeout  time.Duration
	client   hybridWaterHeaterClient
	targets  []echonetlite.Device

	linkageModeForSolarPowerGenerationGauge *prometheus.GaugeVec
	waterHeatingStatusGauge                 *prometheus.GaugeVec
	heaterStatusGauge                       *prometheus.GaugeVec
	hotWaterSupplyStatusGauge               *prometheus.GaugeVec
	remainingWaterGauge                     *prometheus.GaugeVec
	tankCapacityGauge                       *prometheus.GaugeVec
	collectMetrics                          *CollectMetrics
}

type hybridWaterHeaterClient interface {
	Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.HybridWaterHeater, error)
}

func NewHybridWaterHeaterCollector(
	client hybridWaterHeaterClient,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *HybridWaterHeaterCollector {
	return &HybridWaterHeaterCollector{
		interval: interval,
		timeout:  timeout,
		client:   client,
		targets:  targets,
		linkageModeForSolarPowerGenerationGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_hybrid_water_heater_linkage_mode_for_solar_power_generation",
				Help: "Linkage mode setting for solar power generation.",
			},
			[]string{"host", "eoj"},
		),
		waterHeatingStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_hybrid_water_heater_water_heating_status",
				Help: "Water heating status.",
			},
			[]string{"host", "eoj"},
		),
		heaterStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_hybrid_water_heater_heater_status",
				Help: "Heater status.",
			},
			[]string{"host", "eoj"},
		),
		hotWaterSupplyStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_hybrid_water_heater_hot_water_supply_status",
				Help: "Hot water supply status.",
			},
			[]string{"host", "eoj"},
		),
		remainingWaterGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_hybrid_water_heater_remaining_water_cubic_meters",
				Help: "Measured amount of hot water remaining in tank.",
			},
			[]string{"host", "eoj"},
		),
		tankCapacityGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_hybrid_water_heater_tank_capacity_cubic_meters",
				Help: "Tank capacity.",
			},
			[]string{"host", "eoj"},
		),
		collectMetrics: collectMetrics,
	}
}

func (c *HybridWaterHeaterCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *HybridWaterHeaterCollector) Describe(ch chan<- *prometheus.Desc) {
	c.linkageModeForSolarPowerGenerationGauge.Describe(ch)
	c.waterHeatingStatusGauge.Describe(ch)
	c.heaterStatusGauge.Describe(ch)
	c.hotWaterSupplyStatusGauge.Describe(ch)
	c.remainingWaterGauge.Describe(ch)
	c.tankCapacityGauge.Describe(ch)
}

func (c *HybridWaterHeaterCollector) Collect(ch chan<- prometheus.Metric) {
	c.linkageModeForSolarPowerGenerationGauge.Collect(ch)
	c.waterHeatingStatusGauge.Collect(ch)
	c.heaterStatusGauge.Collect(ch)
	c.hotWaterSupplyStatusGauge.Collect(ch)
	c.remainingWaterGauge.Collect(ch)
	c.tankCapacityGauge.Collect(ch)
}

func (c *HybridWaterHeaterCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
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

func (c *HybridWaterHeaterCollector) collect(ctx context.Context, devices []echonetlite.Device) {
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

func (c *HybridWaterHeaterCollector) updateMetrics(device echonetlite.Device, reading *echonetlite.HybridWaterHeater) {
	host, eoj := device.Host(), device.EOJ().String()

	c.linkageModeForSolarPowerGenerationGauge.WithLabelValues(host, eoj).Set(float64(reading.LinkageModeForSolarPowerGeneration))
	for _, status := range []struct {
		gauge *prometheus.GaugeVec
		value *bool
	}{
		{c.waterHeatingStatusGauge, reading.WaterHeatingStatus},
		{c.heaterStatusGauge, reading.HeaterStatus},
		{c.hotWaterSupplyStatusGauge, reading.HotWaterSupplyStatus},
	} {
		if status.value == nil {
			status.gauge.DeleteLabelValues(host, eoj)
			continue
		}
		var value float64
		if *status.value {
			value = 1
		}
		status.gauge.WithLabelValues(host, eoj).Set(value)
	}

	if reading.RemainingWater != nil {
		c.remainingWaterGauge.WithLabelValues(host, eoj).Set(float64(*reading.RemainingWater) / 1_000)
	} else {
		c.remainingWaterGauge.DeleteLabelValues(host, eoj)
	}
	if reading.TankCapacity != nil {
		c.tankCapacityGauge.WithLabelValues(host, eoj).Set(float64(*reading.TankCapacity) / 1_000)
	} else {
		c.tankCapacityGauge.DeleteLabelValues(host, eoj)
	}
}

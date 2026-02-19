package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type storageBatteryGetter interface {
	Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.StorageBattery, error)
}

type StorageBatteryCollector struct {
	interval time.Duration
	timeout  time.Duration
	client   storageBatteryGetter
	targets  []echonetlite.Device

	acChargeableEnergyGauge    *prometheus.GaugeVec
	acDischargeableEnergyGauge *prometheus.GaugeVec
	collectMetrics             *CollectMetrics
}

func NewStorageBatteryCollector(
	conn *echonetlite.Connection,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *StorageBatteryCollector {
	return &StorageBatteryCollector{
		interval: interval,
		timeout:  timeout,
		client:   echonetlite.NewStorageBatteryClient(conn),
		targets:  targets,
		acChargeableEnergyGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_storage_battery_ac_chargeable_electric_energy_joules",
				Help: "AC chargeable electric energy.",
			},
			[]string{"host", "eoj"},
		),
		acDischargeableEnergyGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_storage_battery_ac_dischargeable_electric_energy_joules",
				Help: "AC dischargeable electric energy.",
			},
			[]string{"host", "eoj"},
		),
		collectMetrics: collectMetrics,
	}
}

func (c *StorageBatteryCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *StorageBatteryCollector) Describe(ch chan<- *prometheus.Desc) {
	c.acChargeableEnergyGauge.Describe(ch)
	c.acDischargeableEnergyGauge.Describe(ch)
}

func (c *StorageBatteryCollector) Collect(ch chan<- prometheus.Metric) {
	c.acChargeableEnergyGauge.Collect(ch)
	c.acDischargeableEnergyGauge.Collect(ch)
}

func (c *StorageBatteryCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
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

func (c *StorageBatteryCollector) collect(ctx context.Context, devices []echonetlite.Device) {
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

func (c *StorageBatteryCollector) updateMetrics(
	device echonetlite.Device,
	sb *echonetlite.StorageBattery,
) {
	c.acChargeableEnergyGauge.WithLabelValues(device.Host(), device.EOJ().String()).Set(
		float64(sb.ACChargeableElectricEnergyWh) * 3600.0,
	)
	c.acDischargeableEnergyGauge.WithLabelValues(device.Host(), device.EOJ().String()).Set(
		float64(sb.ACDischargeableElectricEnergyWh) * 3600.0,
	)
}

package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type StorageBatteryCollector struct {
	interval time.Duration
	timeout  time.Duration
	client   storageBatteryClient
	targets  []echonetlite.Device

	acChargeableEnergyGauge    *prometheus.GaugeVec
	acDischargeableEnergyGauge *prometheus.GaugeVec
	acChargingEnergyDesc       *prometheus.Desc
	acChargingEnergy           map[cumulativeStorageBatteryEnergyKey]float64
	acDischargingEnergyDesc    *prometheus.Desc
	acDischargingEnergy        map[cumulativeStorageBatteryEnergyKey]float64
	metricsStateMu             sync.Mutex
	collectMetrics             *CollectMetrics
}

type storageBatteryClient interface {
	Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.StorageBattery, error)
}

type cumulativeStorageBatteryEnergyKey struct {
	host string
	eoj  string
}

func NewStorageBatteryCollector(
	client storageBatteryClient,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *StorageBatteryCollector {
	return &StorageBatteryCollector{
		interval: interval,
		timeout:  timeout,
		client:   client,
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
		acChargingEnergyDesc: prometheus.NewDesc(
			"echonetlite_storage_battery_ac_charging_electric_energy_joules_total",
			"Cumulative charging electric energy (AC).",
			[]string{"host", "eoj"},
			nil,
		),
		acChargingEnergy: make(map[cumulativeStorageBatteryEnergyKey]float64),
		acDischargingEnergyDesc: prometheus.NewDesc(
			"echonetlite_storage_battery_ac_discharging_electric_energy_joules_total",
			"Cumulative discharging electric energy (AC).",
			[]string{"host", "eoj"},
			nil,
		),
		acDischargingEnergy: make(map[cumulativeStorageBatteryEnergyKey]float64),
		collectMetrics:      collectMetrics,
	}
}

func (c *StorageBatteryCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *StorageBatteryCollector) Describe(ch chan<- *prometheus.Desc) {
	c.acChargeableEnergyGauge.Describe(ch)
	c.acDischargeableEnergyGauge.Describe(ch)
	ch <- c.acChargingEnergyDesc
	ch <- c.acDischargingEnergyDesc
}

func (c *StorageBatteryCollector) Collect(ch chan<- prometheus.Metric) {
	c.acChargeableEnergyGauge.Collect(ch)
	c.acDischargeableEnergyGauge.Collect(ch)
	c.metricsStateMu.Lock()
	for key, value := range c.acChargingEnergy {
		ch <- prometheus.MustNewConstMetric(
			c.acChargingEnergyDesc,
			prometheus.CounterValue,
			value,
			key.host,
			key.eoj,
		)
	}
	for key, value := range c.acDischargingEnergy {
		ch <- prometheus.MustNewConstMetric(
			c.acDischargingEnergyDesc,
			prometheus.CounterValue,
			value,
			key.host,
			key.eoj,
		)
	}
	c.metricsStateMu.Unlock()
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
		float64(sb.ACChargeableElectricEnergy) * 3600.0,
	)
	c.acDischargeableEnergyGauge.WithLabelValues(device.Host(), device.EOJ().String()).Set(
		float64(sb.ACDischargeableElectricEnergy) * 3600.0,
	)

	key := cumulativeStorageBatteryEnergyKey{
		host: device.Host(),
		eoj:  device.EOJ().String(),
	}
	c.metricsStateMu.Lock()
	c.acChargingEnergy[key] = float64(sb.ACCumulativeChargingElectricEnergy) * 3600.0
	c.acDischargingEnergy[key] = float64(sb.ACCumulativeDischargingElectricEnergy) * 3600.0
	c.metricsStateMu.Unlock()
}

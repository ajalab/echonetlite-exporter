package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type DRElectricEnergyMeterCollector struct {
	interval time.Duration
	timeout  time.Duration
	client   drElectricEnergyMeterClient
	targets  []echonetlite.Device

	acInOutInstantaneousPowerGauge              *prometheus.GaugeVec
	independentOperationInstantaneousPowerGauge *prometheus.GaugeVec
	acInputCumulativeEnergyDesc                 *prometheus.Desc
	acOutputCumulativeEnergyDesc                *prometheus.Desc
	independentOperationCumulativeEnergyDesc    *prometheus.Desc
	acInputCumulativeEnergy                     map[cumulativeDRElectricEnergyMeterKey]float64
	acOutputCumulativeEnergy                    map[cumulativeDRElectricEnergyMeterKey]float64
	independentOperationCumulativeEnergy        map[cumulativeDRElectricEnergyMeterKey]float64
	metricsStateMu                              sync.Mutex
	collectMetrics                              *CollectMetrics
}

type drElectricEnergyMeterClient interface {
	Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.DRElectricEnergyMeter, error)
}

type cumulativeDRElectricEnergyMeterKey struct {
	host string
	eoj  string
}

func NewDRElectricEnergyMeterCollector(
	client drElectricEnergyMeterClient,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *DRElectricEnergyMeterCollector {
	return &DRElectricEnergyMeterCollector{
		interval: interval,
		timeout:  timeout,
		client:   client,
		targets:  targets,
		acInOutInstantaneousPowerGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_dr_electric_energy_meter_ac_inout_electric_power_watts",
				Help: "Measured instantaneous electric power (AC input/output).",
			},
			[]string{"host", "eoj"},
		),
		independentOperationInstantaneousPowerGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_dr_electric_energy_meter_independent_operation_electric_power_watts",
				Help: "Measured instantaneous electric power (output during a power outage).",
			},
			[]string{"host", "eoj"},
		),
		acInputCumulativeEnergyDesc: prometheus.NewDesc(
			"echonetlite_dr_electric_energy_meter_ac_input_electric_energy_joules_total",
			"Cumulative electric energy (AC input).",
			[]string{"host", "eoj"},
			nil,
		),
		acOutputCumulativeEnergyDesc: prometheus.NewDesc(
			"echonetlite_dr_electric_energy_meter_ac_output_electric_energy_joules_total",
			"Cumulative electric energy (AC output).",
			[]string{"host", "eoj"},
			nil,
		),
		independentOperationCumulativeEnergyDesc: prometheus.NewDesc(
			"echonetlite_dr_electric_energy_meter_independent_operation_electric_energy_joules_total",
			"Cumulative electric energy (output during a power outage).",
			[]string{"host", "eoj"},
			nil,
		),
		acInputCumulativeEnergy:              make(map[cumulativeDRElectricEnergyMeterKey]float64),
		acOutputCumulativeEnergy:             make(map[cumulativeDRElectricEnergyMeterKey]float64),
		independentOperationCumulativeEnergy: make(map[cumulativeDRElectricEnergyMeterKey]float64),
		collectMetrics:                       collectMetrics,
	}
}

func (c *DRElectricEnergyMeterCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *DRElectricEnergyMeterCollector) Describe(ch chan<- *prometheus.Desc) {
	c.acInOutInstantaneousPowerGauge.Describe(ch)
	c.independentOperationInstantaneousPowerGauge.Describe(ch)
	ch <- c.acInputCumulativeEnergyDesc
	ch <- c.acOutputCumulativeEnergyDesc
	ch <- c.independentOperationCumulativeEnergyDesc
}

func (c *DRElectricEnergyMeterCollector) Collect(ch chan<- prometheus.Metric) {
	c.acInOutInstantaneousPowerGauge.Collect(ch)
	c.independentOperationInstantaneousPowerGauge.Collect(ch)

	c.metricsStateMu.Lock()
	for key, value := range c.acInputCumulativeEnergy {
		ch <- prometheus.MustNewConstMetric(c.acInputCumulativeEnergyDesc, prometheus.CounterValue, value, key.host, key.eoj)
	}
	for key, value := range c.acOutputCumulativeEnergy {
		ch <- prometheus.MustNewConstMetric(c.acOutputCumulativeEnergyDesc, prometheus.CounterValue, value, key.host, key.eoj)
	}
	for key, value := range c.independentOperationCumulativeEnergy {
		ch <- prometheus.MustNewConstMetric(c.independentOperationCumulativeEnergyDesc, prometheus.CounterValue, value, key.host, key.eoj)
	}
	c.metricsStateMu.Unlock()
}

func (c *DRElectricEnergyMeterCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
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

func (c *DRElectricEnergyMeterCollector) collect(ctx context.Context, devices []echonetlite.Device) {
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

func (c *DRElectricEnergyMeterCollector) updateMetrics(device echonetlite.Device, m *echonetlite.DRElectricEnergyMeter) {
	c.acInOutInstantaneousPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()).Set(float64(m.ACInstantaneousElectricPower))
	c.independentOperationInstantaneousPowerGauge.WithLabelValues(device.Host(), device.EOJ().String()).Set(float64(m.IndependentOperationInstantaneousElectricPower))

	key := cumulativeDRElectricEnergyMeterKey{host: device.Host(), eoj: device.EOJ().String()}
	scale := float64(m.CumulativeAmountsOfElectricEnergyUnit) * 3600000

	c.metricsStateMu.Lock()
	c.acInputCumulativeEnergy[key] = float64(m.ACInputCumulativeElectricEnergy) * scale
	c.acOutputCumulativeEnergy[key] = float64(m.ACOutputCumulativeElectricEnergy) * scale
	c.independentOperationCumulativeEnergy[key] = float64(m.IndependentOperationCumulativeElectricEnergy) * scale
	c.metricsStateMu.Unlock()
}

package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

type InstantaneousWaterHeaterCollector struct {
	interval time.Duration
	timeout  time.Duration
	client   instantaneousWaterHeaterClient
	targets  []echonetlite.Device

	hotWaterHeatingStatusGauge          *prometheus.GaugeVec
	bathWaterHeatingStatusGauge         *prometheus.GaugeVec
	automaticBathOperationGauge         *prometheus.GaugeVec
	targetSuppliedWaterTemperatureGauge *prometheus.GaugeVec
	targetBathWaterTemperatureGauge     *prometheus.GaugeVec
	showerHotWaterSupplyStatusGauge     *prometheus.GaugeVec
	kitchenHotWaterSupplyStatusGauge    *prometheus.GaugeVec
	bathOperationStatusGauge            *prometheus.GaugeVec
	collectMetrics                      *CollectMetrics
}

type instantaneousWaterHeaterClient interface {
	Get(ctx context.Context, host string, eoj echonetlite.EOJ) (*echonetlite.InstantaneousWaterHeater, error)
}

func NewInstantaneousWaterHeaterCollector(
	client instantaneousWaterHeaterClient,
	interval time.Duration,
	timeout time.Duration,
	collectMetrics *CollectMetrics,
	targets []echonetlite.Device,
) *InstantaneousWaterHeaterCollector {
	return &InstantaneousWaterHeaterCollector{
		interval: interval,
		timeout:  timeout,
		client:   client,
		targets:  targets,
		hotWaterHeatingStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_instantaneous_water_heater_hot_water_heating_status",
				Help: "Hot water heating status.",
			},
			[]string{"host", "eoj"},
		),
		bathWaterHeatingStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_instantaneous_water_heater_bath_water_heating_status",
				Help: "Bath water heater status.",
			},
			[]string{"host", "eoj"},
		),
		automaticBathOperationGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_instantaneous_water_heater_automatic_bath_operation",
				Help: "Bath Auto mode setting.",
			},
			[]string{"host", "eoj"},
		),
		targetSuppliedWaterTemperatureGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_instantaneous_water_heater_target_supplied_water_temperature_celsius",
				Help: "Set value of hot water temperature.",
			},
			[]string{"host", "eoj"},
		),
		targetBathWaterTemperatureGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_instantaneous_water_heater_target_bath_water_temperature_celsius",
				Help: "Set value of bath temperature.",
			},
			[]string{"host", "eoj"},
		),
		showerHotWaterSupplyStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_instantaneous_water_heater_shower_hot_water_supply_status",
				Help: "Shower hot water supply status.",
			},
			[]string{"host", "eoj"},
		),
		kitchenHotWaterSupplyStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_instantaneous_water_heater_kitchen_hot_water_supply_status",
				Help: "Kitchen hot water supply status.",
			},
			[]string{"host", "eoj"},
		),
		bathOperationStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "echonetlite_instantaneous_water_heater_bath_operation_status",
				Help: "Bath operation status monitor.",
			},
			[]string{"host", "eoj"},
		),
		collectMetrics: collectMetrics,
	}
}

func (c *InstantaneousWaterHeaterCollector) Start(ctx context.Context) {
	go c.collectLoop(ctx, c.targets)
}

func (c *InstantaneousWaterHeaterCollector) Describe(ch chan<- *prometheus.Desc) {
	c.hotWaterHeatingStatusGauge.Describe(ch)
	c.bathWaterHeatingStatusGauge.Describe(ch)
	c.automaticBathOperationGauge.Describe(ch)
	c.targetSuppliedWaterTemperatureGauge.Describe(ch)
	c.targetBathWaterTemperatureGauge.Describe(ch)
	c.showerHotWaterSupplyStatusGauge.Describe(ch)
	c.kitchenHotWaterSupplyStatusGauge.Describe(ch)
	c.bathOperationStatusGauge.Describe(ch)
}

func (c *InstantaneousWaterHeaterCollector) Collect(ch chan<- prometheus.Metric) {
	c.hotWaterHeatingStatusGauge.Collect(ch)
	c.bathWaterHeatingStatusGauge.Collect(ch)
	c.automaticBathOperationGauge.Collect(ch)
	c.targetSuppliedWaterTemperatureGauge.Collect(ch)
	c.targetBathWaterTemperatureGauge.Collect(ch)
	c.showerHotWaterSupplyStatusGauge.Collect(ch)
	c.kitchenHotWaterSupplyStatusGauge.Collect(ch)
	c.bathOperationStatusGauge.Collect(ch)
}

func (c *InstantaneousWaterHeaterCollector) collectLoop(ctx context.Context, targets []echonetlite.Device) {
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

func (c *InstantaneousWaterHeaterCollector) collect(ctx context.Context, devices []echonetlite.Device) {
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

func (c *InstantaneousWaterHeaterCollector) updateMetrics(device echonetlite.Device, reading *echonetlite.InstantaneousWaterHeater) {
	host, eoj := device.Host(), device.EOJ().String()

	for _, status := range []struct {
		gauge *prometheus.GaugeVec
		value *bool
	}{
		{c.hotWaterHeatingStatusGauge, &reading.HotWaterHeatingStatus},
		{c.bathWaterHeatingStatusGauge, &reading.BathWaterHeatingStatus},
		{c.automaticBathOperationGauge, reading.AutomaticBathOperation},
		{c.showerHotWaterSupplyStatusGauge, reading.ShowerHotWaterSupplyStatus},
		{c.kitchenHotWaterSupplyStatusGauge, reading.KitchenHotWaterSupplyStatus},
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

	if reading.TargetSuppliedWaterTemperature != nil {
		c.targetSuppliedWaterTemperatureGauge.WithLabelValues(host, eoj).Set(float64(*reading.TargetSuppliedWaterTemperature))
	} else {
		c.targetSuppliedWaterTemperatureGauge.DeleteLabelValues(host, eoj)
	}
	if reading.TargetBathWaterTemperature != nil {
		c.targetBathWaterTemperatureGauge.WithLabelValues(host, eoj).Set(float64(*reading.TargetBathWaterTemperature))
	} else {
		c.targetBathWaterTemperatureGauge.DeleteLabelValues(host, eoj)
	}
	if reading.BathOperationStatus != nil {
		c.bathOperationStatusGauge.WithLabelValues(host, eoj).Set(float64(*reading.BathOperationStatus))
	} else {
		c.bathOperationStatusGauge.DeleteLabelValues(host, eoj)
	}
}

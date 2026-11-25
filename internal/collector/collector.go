package collector

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
)

// collector publishes the last successful collection from each device.
type collector struct {
	interval time.Duration
	targets  []echonetlite.Device
	descs    []metricDescriptor
	read     func(context.Context, echonetlite.Device) ([]prometheus.Metric, error)

	successDescs map[echonetlite.Device]*prometheus.Desc

	mu      sync.RWMutex
	metrics map[echonetlite.Device]deviceMetrics
}

type deviceMetrics struct {
	values  []prometheus.Metric
	success prometheus.Metric
}

type metricReading int

const (
	// gaugeReading covers values such as states and settings without a reading qualifier.
	gaugeReading metricReading = iota
	instantaneousReading
	cumulativeReading
)

// metricDescriptor associates a Prometheus descriptor with the ECHONET Lite property it exports.
type metricDescriptor struct {
	desc              *prometheus.Desc
	propertyShortName string
	unit              string
	reading           metricReading
}

func newPropertyMetricDescriptor(epc uint8, propertyShortName, unit string, reading metricReading, name, propertyNameEN string, labels []string) metricDescriptor {
	return buildMetricDescriptor(epc, propertyShortName, unit, reading, name, propertyNameEN, "", labels)
}

// stateValues is a JSON object mapping exported state codes to MRA descriptions.
func newPropertyMetricDescriptorWithValues(epc uint8, propertyShortName, name, propertyNameEN, stateValues string) metricDescriptor {
	return buildMetricDescriptor(epc, propertyShortName, "", gaugeReading, name, propertyNameEN, stateValues, deviceLabels)
}

func buildMetricDescriptor(epc uint8, propertyShortName, unit string, reading metricReading, name, propertyNameEN, stateValues string, labels []string) metricDescriptor {
	help := propertyNameEN + "."
	if stateValues != "" {
		help += " values=" + stateValues
	}
	help += fmt.Sprintf(" EPC=0x%02X", epc)

	return metricDescriptor{
		desc:              prometheus.NewDesc(name, help, labels, nil),
		propertyShortName: propertyShortName,
		unit:              unit,
		reading:           reading,
	}
}

var deviceLabels = []string{"host", "eoj"}

const (
	joulesPerWattHour     = 3_600
	joulesPerKilowattHour = 3_600_000
)

func newCollector[T any](
	interval time.Duration,
	targets []echonetlite.Device,
	descs []metricDescriptor,
	get func(context.Context, string, echonetlite.EOJ) (*T, error),
	metrics func(echonetlite.Device, *T) []prometheus.Metric,
) *collector {
	successDescs := make(map[echonetlite.Device]*prometheus.Desc, len(targets))
	for _, device := range targets {
		// Fixed device labels give each collector a distinct descriptor for this family.
		successDescs[device] = prometheus.NewDesc(
			"echonetlite_collector_success",
			"Whether the last collection was successful (1 for success, 0 for failure).",
			nil, prometheus.Labels{"host": device.Host(), "eoj": device.EOJ().String()},
		)
	}
	return &collector{
		interval:     interval,
		targets:      targets,
		descs:        descs,
		successDescs: successDescs,
		read: func(ctx context.Context, device echonetlite.Device) ([]prometheus.Metric, error) {
			reading, err := get(ctx, device.Host(), device.EOJ())
			if err != nil {
				return nil, err
			}
			return metrics(device, reading), nil
		},
		metrics: make(map[echonetlite.Device]deviceMetrics),
	}
}

func (c *collector) Start(ctx context.Context) {
	go c.run(ctx)
}

func (c *collector) run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	c.collect(ctx)
	for {
		select {
		case <-ticker.C:
			c.collect(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.successDescs {
		ch <- desc
	}
	for _, desc := range c.descs {
		ch <- desc.desc
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	var metrics []prometheus.Metric
	for _, deviceMetrics := range c.metrics {
		metrics = append(metrics, deviceMetrics.success)
		metrics = append(metrics, deviceMetrics.values...)
	}
	c.mu.RUnlock()

	for _, metric := range metrics {
		ch <- metric
	}
}

func (c *collector) collect(ctx context.Context) {
	for _, device := range c.targets {
		if ctx.Err() != nil {
			return
		}
		metrics, err := c.read(ctx, device)
		if ctx.Err() != nil {
			return
		}
		c.mu.Lock()
		state := c.metrics[device]
		if err == nil {
			state.values = metrics
		}
		state.success = prometheus.MustNewConstMetric(c.successDescs[device], prometheus.GaugeValue, boolValue(err == nil))
		c.metrics[device] = state
		c.mu.Unlock()
		if err != nil {
			slog.Warn("failed to collect stats", "host", device.Host(), "eoj", device.EOJ().String(), "err", err)
		}
	}
}

func boolValue(on bool) float64 {
	if on {
		return 1
	}
	return 0
}

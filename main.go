package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ajalab/echonetlite-exporter/internal/collector"
	"github.com/ajalab/echonetlite-exporter/internal/echonetlite"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	webListenAddr         = flag.String("web.listenAddr", ":9200", "Address to listen on for HTTP requests")
	netMulticastInterface = flag.String("net.multicastInterface", "", "Network interface for UDP multicast")
	discoveryScanDuration = flag.Duration("discovery.scanDuration", 10*time.Second, "Duration for scanning ECHONET Lite nodes")
	collectorInterval     = flag.Duration("collector.interval", 15*time.Second, "Interval for collecting metrics from ECHONET Lite devices")
	collectorTimeout      = flag.Duration("collector.timeout", 10*time.Second, "Timeout for collecting metrics from ECHONET Lite devices")
)

type deviceCollector interface {
	prometheus.Collector
	Start(context.Context)
}

type Exporter struct {
	collectors     []deviceCollector
	collectMetrics *collector.CollectMetrics
}

func NewExporter(conn *echonetlite.Connection, targets map[uint16][]echonetlite.Device) *Exporter {
	collectMetrics := collector.NewCollectMetrics()
	interval, timeout := *collectorInterval, *collectorTimeout
	return &Exporter{
		collectMetrics: collectMetrics,
		collectors: []deviceCollector{
			collector.NewPowerDistributionBoardMeteringCollector(
				echonetlite.NewPowerDistributionBoardMeteringClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassPowerDistributionBoardMetering],
			),
			collector.NewDRElectricEnergyMeterCollector(
				echonetlite.NewDRElectricEnergyMeterClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassDRElectricEnergyMeter],
			),
			collector.NewPVPowerGenerationCollector(
				echonetlite.NewPVPowerGenerationClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassPVPowerGeneration],
			),
			collector.NewStorageBatteryCollector(
				echonetlite.NewStorageBatteryClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassStorageBattery],
			),
			collector.NewMultipleInputPCSCollector(
				echonetlite.NewMultipleInputPCSClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassMultipleInputPCS],
			),
			collector.NewInstantaneousWaterHeaterCollector(
				echonetlite.NewInstantaneousWaterHeaterClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassInstantaneousWaterHeater],
			),
			collector.NewWattHourMeterCollector(
				echonetlite.NewWattHourMeterClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassWattHourMeter],
			),
			collector.NewWaterFlowMeterCollector(
				echonetlite.NewWaterFlowMeterClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassWaterFlowMeter],
			),
			collector.NewGasMeterCollector(
				echonetlite.NewGasMeterClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassGasMeter],
			),
			collector.NewHybridWaterHeaterCollector(
				echonetlite.NewHybridWaterHeaterClient(conn),
				interval, timeout, collectMetrics, targets[echonetlite.ClassHybridWaterHeater],
			),
		},
	}
}

func (e *Exporter) RegisterMetrics() {
	prometheus.MustRegister(e.collectMetrics.Collector())
	for _, c := range e.collectors {
		prometheus.MustRegister(c)
	}
}

func (e *Exporter) Start(ctx context.Context) {
	for _, c := range e.collectors {
		c.Start(ctx)
	}
}

func scanTargets(ctx context.Context, conn *echonetlite.Connection) (map[uint16][]echonetlite.Device, error) {
	reqCtx, cancel := context.WithTimeout(ctx, *discoveryScanDuration)
	defer cancel()

	nodeProfileClient := echonetlite.NewNodeProfileClient(conn)

	slog.Info("scanning ECHONET Lite nodes")
	nodes, err := nodeProfileClient.Scan(reqCtx)
	if err != nil {
		return nil, err
	}
	slog.Info(fmt.Sprintf("scanned %d ECHONET Lite nodes", len(nodes)))

	targets := make(map[uint16][]echonetlite.Device)
	for _, node := range nodes {
		for _, eoj := range node.NodeProfile.SelfNodeInstanceListS {
			class := eoj.Class()
			targets[class] = append(targets[class], echonetlite.NewDevice(node.Host, eoj))
		}
	}
	return targets, nil
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: *webListenAddr}

	go func() {
		slog.Info(fmt.Sprintf("listening on %s", *webListenAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	conn, err := echonetlite.NewConnection(*netMulticastInterface)
	if err != nil {
		log.Fatalf("connection error: %v", err)
	}
	defer conn.Close()

	targets, err := scanTargets(ctx, conn)
	if err != nil {
		log.Fatalf("startup failed: failed to scan nodes: %v", err)
	}

	exporter := NewExporter(conn, targets)
	exporter.RegisterMetrics()
	exporter.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown error", "err", err)
	}
}

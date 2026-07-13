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

type Exporter struct {
	conn *echonetlite.Connection

	pdbmCollector                  *collector.PowerDistributionBoardMeteringCollector
	drElectricEnergyMeterCollector *collector.DRElectricEnergyMeterCollector
	pvCollector                    *collector.PVPowerGenerationCollector
	storageBatteryCollector        *collector.StorageBatteryCollector
	multipleInputPCSCollector      *collector.MultipleInputPCSCollector

	collectMetrics *collector.CollectMetrics
}

func NewExporter(
	conn *echonetlite.Connection,
	pdbmTargets []echonetlite.Device,
	drElectricEnergyMeterTargets []echonetlite.Device,
	pvTargets []echonetlite.Device,
	storageBatteryTargets []echonetlite.Device,
	multipleInputPCSTargets []echonetlite.Device,
) *Exporter {
	collectMetrics := collector.NewCollectMetrics()
	pdbmClient := echonetlite.NewPowerDistributionBoardMeteringClient(conn)
	drElectricEnergyMeterClient := echonetlite.NewDRElectricEnergyMeterClient(conn)
	pvClient := echonetlite.NewPVPowerGenerationClient(conn)
	storageBatteryClient := echonetlite.NewStorageBatteryClient(conn)
	multipleInputPCSClient := echonetlite.NewMultipleInputPCSClient(conn)
	pdbmCollector := collector.NewPowerDistributionBoardMeteringCollector(
		pdbmClient,
		*collectorInterval,
		*collectorTimeout,
		collectMetrics,
		pdbmTargets,
	)
	pvCollector := collector.NewPVPowerGenerationCollector(
		pvClient,
		*collectorInterval,
		*collectorTimeout,
		collectMetrics,
		pvTargets,
	)
	drElectricEnergyMeterCollector := collector.NewDRElectricEnergyMeterCollector(
		drElectricEnergyMeterClient,
		*collectorInterval,
		*collectorTimeout,
		collectMetrics,
		drElectricEnergyMeterTargets,
	)
	storageBatteryCollector := collector.NewStorageBatteryCollector(
		storageBatteryClient,
		*collectorInterval,
		*collectorTimeout,
		collectMetrics,
		storageBatteryTargets,
	)
	multipleInputPCSCollector := collector.NewMultipleInputPCSCollector(
		multipleInputPCSClient,
		*collectorInterval,
		*collectorTimeout,
		collectMetrics,
		multipleInputPCSTargets,
	)
	exporter := &Exporter{
		conn:                           conn,
		pdbmCollector:                  pdbmCollector,
		drElectricEnergyMeterCollector: drElectricEnergyMeterCollector,
		pvCollector:                    pvCollector,
		storageBatteryCollector:        storageBatteryCollector,
		multipleInputPCSCollector:      multipleInputPCSCollector,
		collectMetrics:                 collectMetrics,
	}

	return exporter
}

func (e *Exporter) RegisterMetrics() {
	prometheus.MustRegister(
		e.collectMetrics.Collector(),
		e.pdbmCollector,
		e.drElectricEnergyMeterCollector,
		e.pvCollector,
		e.storageBatteryCollector,
		e.multipleInputPCSCollector,
	)
}

func (e *Exporter) Start(ctx context.Context) {
	e.pdbmCollector.Start(ctx)
	e.drElectricEnergyMeterCollector.Start(ctx)
	e.pvCollector.Start(ctx)
	e.storageBatteryCollector.Start(ctx)
	e.multipleInputPCSCollector.Start(ctx)
}

func scanTargets(
	ctx context.Context,
	conn *echonetlite.Connection,
) ([]echonetlite.Device, []echonetlite.Device, []echonetlite.Device, []echonetlite.Device, []echonetlite.Device, error) {
	reqCtx, cancel := context.WithTimeout(ctx, *discoveryScanDuration)
	defer cancel()

	nodeProfileClient := echonetlite.NewNodeProfileClient(conn)

	slog.Info("scanning ECHONET Lite nodes")
	nodes, err := nodeProfileClient.Scan(reqCtx)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	slog.Info(fmt.Sprintf("scanned %d ECHONET Lite nodes", len(nodes)))

	var pdbms []echonetlite.Device
	var drems []echonetlite.Device
	var pvpgs []echonetlite.Device
	var sbs []echonetlite.Device
	var mipcss []echonetlite.Device
	for _, node := range nodes {
		nodeProfile := node.NodeProfile
		for _, eoj := range nodeProfile.SelfNodeInstanceListS {
			class := eoj.Class()
			switch class {
			case echonetlite.ClassPowerDistributionBoardMetering:
				pdbms = append(pdbms, echonetlite.NewDevice(node.Host, eoj))
			case echonetlite.ClassDRElectricEnergyMeter:
				drems = append(drems, echonetlite.NewDevice(node.Host, eoj))
			case echonetlite.ClassPVPowerGeneration:
				pvpgs = append(pvpgs, echonetlite.NewDevice(node.Host, eoj))
			case echonetlite.ClassStorageBattery:
				sbs = append(sbs, echonetlite.NewDevice(node.Host, eoj))
			case echonetlite.ClassMultipleInputPCS:
				mipcss = append(mipcss, echonetlite.NewDevice(node.Host, eoj))
			}
		}
	}

	return pdbms, drems, pvpgs, sbs, mipcss, nil
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

	pdbms, drems, pvs, sbs, mipcss, err := scanTargets(ctx, conn)
	if err != nil {
		log.Fatalf("startup failed: failed to scan nodes: %v", err)
	}

	exporter := NewExporter(conn, pdbms, drems, pvs, sbs, mipcss)
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

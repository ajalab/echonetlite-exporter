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

	"echonetlite-exporter/collector"
	"echonetlite-exporter/echonetlite"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	webListenAddr         = flag.String("net.listenAddr", ":9200", "Address to listen on for HTTP requests")
	netMulticastInterface = flag.String("net.multicastInterface", "", "Network interface for UDP multicast")
	scannerTimeout        = flag.Duration("scanner.timeout", 10*time.Second, "Timeout for scanning ECHONET Lite nodes")
	collectorInterval     = flag.Duration("collector.interval", 15*time.Second, "Interval for collecting metrics from nodes")
	collectorTimeout      = flag.Duration("collector.timeout", 10*time.Second, "Timeout for collecting metrics from nodes")
)

type Exporter struct {
	conn    *echonetlite.Connection
	scanner echonetlite.Scanner

	pdbmCollector *collector.PowerDistributionBoardMeteringCollector
	pvCollector   *collector.PVPowerGenerationCollector

	collectMetrics *collector.CollectMetrics
}

func NewExporter(conn *echonetlite.Connection) *Exporter {
	collectMetrics := collector.NewCollectMetrics()
	pdbmCollector := collector.NewPowerDistributionBoardMeteringCollector(
		conn,
		*collectorInterval,
		*collectorTimeout,
		collectMetrics,
	)
	pvCollector := collector.NewPVPowerGenerationCollector(
		conn,
		*collectorInterval,
		*collectorTimeout,
		collectMetrics,
	)
	exporter := &Exporter{
		conn:           conn,
		scanner:        echonetlite.NewScanner(conn),
		pdbmCollector:  pdbmCollector,
		pvCollector:    pvCollector,
		collectMetrics: collectMetrics,
	}

	return exporter
}

func (e *Exporter) RegisterMetrics() {
	prometheus.MustRegister(
		e.collectMetrics.Collector(),
		e.pdbmCollector,
		e.pvCollector,
	)
}

func (e *Exporter) Start(ctx context.Context) error {
	pdbms, pvs, err := e.scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to scan nodes: %v", err)
	}

	e.pdbmCollector.Start(ctx, pdbms)
	e.pvCollector.Start(ctx, pvs)
	return nil
}

func (e *Exporter) scan(ctx context.Context) ([]*echonetlite.PowerDistributionBoardMetering, []*echonetlite.PVPowerGeneration, error) {
	reqCtx, cancel := context.WithTimeout(ctx, *scannerTimeout)
	defer cancel()

	slog.Info("scanning ECHONET Lite nodes")
	nodes, err := e.scanner.Scan(reqCtx)
	if err != nil {
		return nil, nil, err
	}
	slog.Info(fmt.Sprintf("scanned %d ECHONET Lite nodes", len(nodes)))

	var pdbms []*echonetlite.PowerDistributionBoardMetering
	var pvs []*echonetlite.PVPowerGeneration
	for _, node := range nodes {
		nodeProfile := node.NodeProfile
		for _, eoj := range nodeProfile.SelfNodeInstanceListS {
			if eoj[0] == 0x02 && eoj[1] == 0x87 {
				pdbms = append(pdbms, echonetlite.NewPowerDistributionBoardMetering(
					node.Host,
					eoj,
					e.conn,
				),
				)
			}
			if eoj[0] == 0x02 && eoj[1] == 0x79 {
				pvs = append(pvs, echonetlite.NewPVPowerGeneration(
					node.Host,
					eoj,
					e.conn,
				),
				)
			}
		}
	}

	return pdbms, pvs, nil
}

func main() {
	flag.Parse()

	conn, err := echonetlite.NewConnection(*netMulticastInterface)
	if err != nil {
		log.Fatalf("connection error: %v", err)
	}
	defer conn.Close()

	exporter := NewExporter(conn)
	exporter.RegisterMetrics()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := exporter.Start(ctx); err != nil {
		log.Fatalf("startup failed: %v", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: *webListenAddr}

	go func() {
		slog.Info(fmt.Sprintf("listening on %s", *webListenAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown error", "err", err)
	}
}

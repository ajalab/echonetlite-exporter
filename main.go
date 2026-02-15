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
	netMulticastInterface = flag.String("net.multicastInterface", "", "Network interface for UDP multicast")
	scannerInterval       = flag.Duration("scanner.interval", 60*time.Second, "Interval for scanning ECHONET Lite nodes")
	scannerTimeout        = flag.Duration("scanner.timeout", 10*time.Second, "Timeout for scanning ECHONET Lite nodes")
	collectorInterval     = flag.Duration("collector.interval", 15*time.Second, "Interval for collecting metrics from nodes")
	collectorTimeout      = flag.Duration("collector.timeout", 10*time.Second, "Timeout for collecting metrics from nodes")
)

const (
	listenAddr = ":9200"
)

type Exporter struct {
	conn    *echonetlite.Connection
	scanner echonetlite.Scanner

	scanErrorsTotal prometheus.Counter

	pdbmCollector *collector.PowerDistributionBoardMeteringCollector
	pdbmUpdates   chan []*echonetlite.PowerDistributionBoardMetering
	pvCollector   *collector.PVPowerGenerationCollector
	pvUpdates     chan []*echonetlite.PVPowerGeneration

	collectMetrics *collector.CollectMetrics
}

func NewExporter(conn *echonetlite.Connection) *Exporter {
	pdbmUpdates := make(chan []*echonetlite.PowerDistributionBoardMetering)
	collectMetrics := collector.NewCollectMetrics()
	pdbmCollector := collector.NewPowerDistributionBoardMeteringCollector(
		conn,
		pdbmUpdates,
		*collectorInterval,
		*collectorTimeout,
		collectMetrics,
	)
	pvUpdates := make(chan []*echonetlite.PVPowerGeneration)
	pvCollector := collector.NewPVPowerGenerationCollector(
		conn,
		pvUpdates,
		*collectorInterval,
		*collectorTimeout,
		collectMetrics,
	)
	exporter := &Exporter{
		conn:    conn,
		scanner: echonetlite.NewScanner(conn),
		scanErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "echonetlite_exporter_scan_errors_total",
			Help: "Total number of scan errors.",
		}),
		pdbmCollector:  pdbmCollector,
		pdbmUpdates:    pdbmUpdates,
		pvCollector:    pvCollector,
		pvUpdates:      pvUpdates,
		collectMetrics: collectMetrics,
	}

	return exporter
}

func (e *Exporter) RegisterMetrics() {
	prometheus.MustRegister(
		e.scanErrorsTotal,
		e.collectMetrics.Collector(),
		e.pdbmCollector,
		e.pvCollector,
	)
}

func (e *Exporter) Start(ctx context.Context) {
	go e.scanLoop(ctx)
	e.pdbmCollector.Start(ctx)
	e.pvCollector.Start(ctx)
}

func (e *Exporter) scanLoop(ctx context.Context) {
	ticker := time.NewTicker(*scannerInterval)
	defer ticker.Stop()

	e.runScan(ctx)
	for {
		select {
		case <-ticker.C:
			e.runScan(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (e *Exporter) runScan(ctx context.Context) {
	reqCtx, cancel := context.WithTimeout(ctx, *scannerTimeout)
	defer cancel()

	nodes, err := e.scanner.ScanNodes(reqCtx)
	if err != nil {
		e.scanErrorsTotal.Inc()
		slog.Error("failed to scan nodes", "err", err)
		return
	}

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
	e.pdbmUpdates <- pdbms
	e.pvUpdates <- pvs
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

	exporter.Start(ctx)

	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: listenAddr}

	go func() {
		slog.Info(fmt.Sprintf("listening on %s", listenAddr))
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

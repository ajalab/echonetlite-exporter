package main

import (
	"context"
	"log"
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

const (
	scanInterval   = 60 * time.Second
	requestTimeout = 10 * time.Second
	listenAddr     = ":9100"
)

type Exporter struct {
	conn    *echonetlite.Connection
	scanner echonetlite.Scanner

	scanErrorsTotal prometheus.Counter

	pdbmCollector *collector.PowerDistributionBoardMeterCollector
	pdbmUpdates   chan []*echonetlite.PowerDistributionBoardMetering
}

func NewExporter(conn *echonetlite.Connection) *Exporter {
	pdbmUpdates := make(chan []*echonetlite.PowerDistributionBoardMetering)
	pdbmCollector := collector.NewPowerDistributionBoardMeterCollector(conn, pdbmUpdates)
	exporter := &Exporter{
		conn:    conn,
		scanner: echonetlite.NewScanner(conn),
		scanErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "echonetlite_exporter_scan_errors_total",
			Help: "Total number of scan errors.",
		}),
		pdbmCollector: pdbmCollector,
		pdbmUpdates:   pdbmUpdates,
	}

	return exporter
}

func (e *Exporter) RegisterMetrics() {
	prometheus.MustRegister(
		e.scanErrorsTotal,
		e.pdbmCollector,
	)
}

func (e *Exporter) Start(ctx context.Context) {
	go e.scanLoop(ctx)
	e.pdbmCollector.Start(ctx)
}

func (e *Exporter) scanLoop(ctx context.Context) {
	ticker := time.NewTicker(scanInterval)
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
	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	nodes, err := e.scanner.ScanNodes(reqCtx)
	if err != nil {
		e.scanErrorsTotal.Inc()
		log.Printf("scan error: %v", err)
		return
	}

	var pdbms []*echonetlite.PowerDistributionBoardMetering
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
		}
	}
	e.pdbmUpdates <- pdbms
}

func main() {
	conn, err := echonetlite.NewConnection()
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
		log.Printf("listening on %s", listenAddr)
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
		log.Printf("http shutdown error: %v", err)
	}
}

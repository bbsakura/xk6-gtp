// Command pgw is a reference GTPv2-C P-GW used by the k6/x/gtpv2 examples and
// tests. All handler logic lives in pkg/refpgw so this binary stays a thin
// wrapper around configuration + serve loop.
package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wmnsk/go-gtp/gtpv2"

	"github.com/bbsakura/xk6-gtp/pkg/refpgw"
)

var (
	s5c        = flag.String("s5c", "127.0.0.1", "IP for S5-C interface.")
	s5u        = flag.String("s5u", "127.0.0.1", "IP for S5-U interface.")
	s5cport    = flag.String("cport", gtpv2.GTPCPort, "GTP CPort")
	logFormat  = flag.String("log-format", "text", "Log output format: text or json.")
	logLevel   = flag.String("log-level", "info", "Log verbosity: debug, info, warn, error.")
	reportEach = flag.Duration("report-interval", 100*time.Second, "How often to log the active subscriber list. Set to 0 to disable.")
)

func main() {
	flag.Parse()

	logger := newLogger(*logFormat, *logLevel).With("component", "refpgw")

	s5cAddr, err := net.ResolveUDPAddr("udp", *s5c+*s5cport)
	if err != nil {
		logger.Error("resolve S5-C addr", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	conn := gtpv2.NewConn(s5cAddr, gtpv2.IFTypeS5S8PGWGTPC, 0)
	handlers := refpgw.New(refpgw.Config{
		S5UAddr: *s5u,
		Logger:  logger,
	})
	handlers.AddTo(conn)

	go func() {
		if err := conn.ListenAndServe(ctx); err != nil && ctx.Err() == nil {
			logger.Error("serve failed", "error", err)
		}
	}()
	logger.Info("started serving C-Plane", "addr", s5cAddr.String())

	reportActiveSubscribers(ctx, logger, conn, *reportEach)
}

func newLogger(format, level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	var h slog.Handler
	switch format {
	case "json":
		h = slog.NewJSONHandler(os.Stderr, opts)
	default:
		h = slog.NewTextHandler(os.Stderr, opts)
	}
	return slog.New(h)
}

func parseLevel(s string) slog.Level {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(s)); err == nil {
		return lvl
	}
	return slog.LevelInfo
}

// reportActiveSubscribers periodically logs which subscribers currently hold
// an active session. Runs until ctx is cancelled.
func reportActiveSubscribers(ctx context.Context, logger *slog.Logger, conn *gtpv2.Conn, interval time.Duration) {
	if interval <= 0 {
		<-ctx.Done()
		return
	}
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			var imsis []string
			for _, s := range conn.Sessions() {
				if !s.IsActive() {
					continue
				}
				imsis = append(imsis, s.IMSI)
			}
			if len(imsis) == 0 {
				continue
			}
			logger.Info("active subscribers", "count", len(imsis), "imsis", imsis)
		}
	}
}

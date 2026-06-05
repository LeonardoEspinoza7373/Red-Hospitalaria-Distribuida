package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/proxy"
)

func main() {
	httpPort := flag.String("http-port", "8080", "HTTP proxy port")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	p := proxy.New(*httpPort)
	if err := p.Start(); err != nil {
		slog.Error("failed to start proxy", "error", err)
		os.Exit(1)
	}

	ap := proxy.NewAccessPoint()
	if err := ap.Start(); err != nil {
		slog.Error("failed to start AP", "error", err)
		os.Exit(1)
	}

	slog.Info("proxy ready", "coordinator_snoop", ":5000", "http_proxy", ":"+*httpPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutting down")
	ap.Stop()
	p.Stop()
}

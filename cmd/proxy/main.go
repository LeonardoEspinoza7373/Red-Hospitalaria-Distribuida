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
	listenAddr := flag.String("listen", ":80", "Dirección HTTP (ej :80, 192.168.2.1:80)")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	p := proxy.New(*listenAddr)
	if err := p.Start(); err != nil {
		slog.Error("failed to start proxy", "error", err)
		os.Exit(1)
	}

	slog.Info("proxy ready", "coordinator_snoop", ":5000", "http_proxy", *listenAddr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutting down")
	p.Stop()
}

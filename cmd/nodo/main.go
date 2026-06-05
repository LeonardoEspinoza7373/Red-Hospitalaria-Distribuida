package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/auth"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/data"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/node"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/pkg/config"
)

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			ip := ipnet.IP.String()
			if _, exists := config.IPToID[ip]; exists {
				return ip
			}
		}
	}
	return ""
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	ip := getLocalIP()
	if ip == "" {
		slog.Error("no se pudo detectar una IP válida (192.168.1.10-13)")
		slog.Info("asegúrate de estar conectado a la red RedHospitalaria")
		os.Exit(1)
	}

	n := node.New(ip)
	if n == nil {
		slog.Error("IP detectada no está en la configuración", "ip", ip)
		os.Exit(1)
	}

	n.SetHTTPAddr(":" + config.FrontendPort)

	dataDir := filepath.Join(".", "data", fmt.Sprintf("%d", n.ID))
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		slog.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	store := data.NewUserStore(filepath.Join(dataDir, "users.json"))
	if err := store.Load(); err != nil {
		slog.Error("failed to load users", "error", err)
		os.Exit(1)
	}

	n.UserStore = store
	n.SessionStore = auth.NewSessionStore(24 * time.Hour)

	users, err := store.List()
	if err != nil {
		slog.Error("failed to list users", "error", err)
		os.Exit(1)
	}
	if len(users) == 0 {
		admin := &data.User{
			Username:    "admin",
			Password:    auth.HashPassword("admin"),
			DisplayName: "Administrador",
			Role:        "admin",
			HospitalID:  n.ID,
		}
		if err := store.Create(admin); err != nil {
			slog.Error("failed to create admin user", "error", err)
			os.Exit(1)
		}
		slog.Info("default admin user created (admin / admin)")
	}

	n.Start()

	slog.Info("nodo listo", "node_id", n.ID, "ip", ip)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("deteniendo nodo")
	n.Stop()
}

package proxy

import (
	"log/slog"
)

type AccessPoint struct {
	log *slog.Logger
}

func NewAccessPoint() *AccessPoint {
	return &AccessPoint{
		log: slog.With("component", "ap"),
	}
}

func (ap *AccessPoint) Start() error {
	ap.log.Info("AP manager initialized (hostapd/dnsmasq control pending)")
	return nil
}

func (ap *AccessPoint) Stop() {
	ap.log.Info("AP manager stopped")
}

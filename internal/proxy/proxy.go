package proxy

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/transport"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/pkg/config"
)

type Proxy struct {
	mu              sync.Mutex
	coordinatorID   int
	coordinatorAddr string
	lastHeartbeat   time.Time

	tcpAddr    string
	httpAddr   string
	server     *transport.Server
	httpServer *http.Server
	rp         *httputil.ReverseProxy
	ctx        context.Context
	cancel     context.CancelFunc
	log        *slog.Logger

	nodeAddrs map[int]string
}

func New(httpAddr string) *Proxy {
	return NewWithPorts(config.Port, httpAddr)
}

func NewWithPorts(tcpPort, httpAddr string) *Proxy {
	nodeAddrs := make(map[int]string)
	for id, ip := range config.IDToIP {
		nodeAddrs[id] = net.JoinHostPort(ip, config.FrontendPort)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// If httpAddr is just a port number (no colon), prefix with ":"
	if httpAddr != "" && httpAddr[0] != ':' && httpAddr[0] != '@' {
		httpAddr = ":" + httpAddr
	}

	p := &Proxy{
		tcpAddr:   ":" + tcpPort,
		httpAddr:  httpAddr,
		ctx:       ctx,
		cancel:    cancel,
		log:       slog.With("component", "proxy"),
		nodeAddrs: nodeAddrs,
	}

	p.rp = &httputil.ReverseProxy{
		Director: p.directRequest,
	}

	return p
}

func (p *Proxy) Start() error {
	p.log.Info("starting proxy", "tcp", p.tcpAddr, "http", p.httpAddr)

	p.server = transport.NewServer(p.tcpAddr, p.onMessage)
	if err := p.server.Start(); err != nil {
		return err
	}

	p.httpServer = &http.Server{
		Addr:    p.httpAddr,
		Handler: http.HandlerFunc(p.serveHTTP),
	}

	go func() {
		p.log.Info("HTTP proxy listening", "addr", p.httpAddr)
		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.log.Error("HTTP server error", "error", err)
		}
	}()

	go p.healthLoop()

	return nil
}

func (p *Proxy) Stop() {
	p.log.Info("stopping proxy")
	p.cancel()

	if p.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p.httpServer.Shutdown(shutdownCtx)
	}

	if p.server != nil {
		p.server.Stop()
	}
}

func (p *Proxy) onMessage(msg protocol.Message, addr net.Addr) {
	switch msg.Type {
	case protocol.Coordinator:
		p.log.Info("discovered coordinator via COORDINATOR", "coordinator_id", msg.CoordinatorID)
		p.setCoordinator(msg.CoordinatorID)
	case protocol.Heartbeat:
		p.log.Debug("heartbeat from coordinator", "coordinator_id", msg.CoordinatorID)
		p.setCoordinator(msg.CoordinatorID)
	}
}

func (p *Proxy) setCoordinator(id int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if id != p.coordinatorID {
		p.log.Info("coordinator changed", "old", p.coordinatorID, "new", id)
		p.coordinatorID = id
		if addr, ok := p.nodeAddrs[id]; ok {
			p.coordinatorAddr = addr
		}
	}
	p.lastHeartbeat = time.Now()
}

func (p *Proxy) getCoordinatorAddr() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.coordinatorAddr
}

func (p *Proxy) serveHTTP(w http.ResponseWriter, r *http.Request) {
	addr := p.getCoordinatorAddr()
	if addr == "" {
		http.Error(w, "no coordinator available", http.StatusServiceUnavailable)
		return
	}
	r.URL.Host = addr
	r.URL.Scheme = "http"
	p.rp.ServeHTTP(w, r)
}

func (p *Proxy) directRequest(r *http.Request) {
	addr := p.getCoordinatorAddr()
	if addr == "" {
		return
	}
	r.URL.Scheme = "http"
	r.URL.Host = addr
	r.Host = addr
}

func (p *Proxy) healthLoop() {
	ticker := time.NewTicker(config.HeartbeatTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.mu.Lock()
			stale := p.coordinatorID != 0 && time.Since(p.lastHeartbeat) > config.HeartbeatTimeout
			p.mu.Unlock()

			if stale {
				p.log.Warn("coordinator heartbeat stale, clearing")
				p.clearCoordinator()
			}
		}
	}
}

func (p *Proxy) clearCoordinator() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.coordinatorID = 0
	p.coordinatorAddr = ""
}

func (p *Proxy) CurrentCoordinator() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.coordinatorID
}

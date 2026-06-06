package node

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/auth"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/data"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/transport"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/pkg/config"
)

type NodeState int

const (
	Follower    NodeState = iota
	Candidate
	Coordinator
)

type EventType int

const (
	MsgReceived          EventType = iota
	HeartbeatWatchTimeout
	ElectionResponseTimeout
	StartupElection
)

type Event struct {
	Type    EventType
	Msg     *protocol.Message
	FromAddr string
}

type Node struct {
	ID             int
	IP             string
	Port           string
	Peers          map[int]string
	ProxyAddr      string
	State          NodeState
	CoordinatorID  int

	mu                  sync.Mutex
	ctx                 context.Context
	cancel              context.CancelFunc
	server              *transport.Server
	events              chan Event
	gotOK               bool
	log                 *slog.Logger

	heartbeatWatchTimer *time.Timer
	heartbeatCancel     context.CancelFunc
	electionSeq         int64

	httpServer     *http.Server
	httpAddr       string // empty = do not start HTTP server
	frontendDir    string // path to built SPA (frontend/dist)

	UserStore    *data.UserStore
	SessionStore *auth.SessionStore
}

func New(ip string) *Node {
	return NewWithPort(ip, config.Port)
}

func NewWithPort(ip, port string) *Node {
	return NewWithPortAndPeers(ip, port, nil)
}

func NewWithPortAndPeers(ip, port string, peers map[int]string) *Node {
	id, ok := config.IPToID[ip]
	if !ok {
		slog.Error("unknown IP address", "ip", ip)
		return nil
	}

	if peers == nil {
		peers = make(map[int]string)
		for peerID, peerIP := range config.IDToIP {
			if peerID != id {
				peers[peerID] = net.JoinHostPort(peerIP, config.Port)
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Node{
		ID:        id,
		IP:        ip,
		Port:      port,
		Peers:     peers,
		ProxyAddr: config.ProxyAddr,
		State:     Follower,
		ctx:       ctx,
		cancel:    cancel,
		events:    make(chan Event, 64),
		log:       slog.With("node_id", id, "ip", ip),
	}
}

func (n *Node) SetHTTPAddr(addr string) {
	n.httpAddr = addr
}

func (n *Node) SetFrontendDir(dir string) {
	n.frontendDir = dir
}

func (n *Node) Start() {
	n.log.Info("starting node", "port", n.Port)

	n.server = transport.NewServer(
		net.JoinHostPort(n.IP, n.Port),
		n.onMessage,
	)

	if err := n.server.Start(); err != nil {
		n.log.Error("failed to start TCP server", "error", err)
		return
	}

	go n.eventLoop()
	go n.scheduleStartupElection()
}

func (n *Node) Stop() {
	n.log.Info("stopping node")
	n.cancel()
	if n.heartbeatWatchTimer != nil {
		n.heartbeatWatchTimer.Stop()
	}
	n.mu.Lock()
	if n.heartbeatCancel != nil {
		n.heartbeatCancel()
	}
	n.mu.Unlock()
	n.stopHTTPServer()
	if n.server != nil {
		n.server.Stop()
	}
}

func (n *Node) onMessage(msg protocol.Message, addr net.Addr) {
	select {
	case n.events <- Event{Type: MsgReceived, Msg: &msg, FromAddr: addr.String()}:
	case <-n.ctx.Done():
	}
}

func (n *Node) eventLoop() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case evt := <-n.events:
			n.handleEvent(evt)
		}
	}
}

func (n *Node) handleEvent(evt Event) {
	switch evt.Type {
	case MsgReceived:
		if evt.Msg != nil {
			n.dispatchMessage(*evt.Msg, evt.FromAddr)
		}
	case HeartbeatWatchTimeout:
		n.handleHeartbeatWatchTimeout()
	case ElectionResponseTimeout:
		n.handleElectionTimeout()
	case StartupElection:
		n.handleStartupElection()
	}
}

func (n *Node) dispatchMessage(msg protocol.Message, fromAddr string) {
	switch msg.Type {
	case protocol.Heartbeat:
		n.log.Debug("recv HEARTBEAT", "from", msg.NodeID, "coord", msg.CoordinatorID)
		n.handleHeartbeat(msg.NodeID, msg.CoordinatorID)
	case protocol.Election:
		n.log.Debug("recv ELECTION", "from", msg.NodeID)
		n.handleElection(msg.NodeID, fromAddr)
	case protocol.OK:
		n.log.Debug("recv OK", "from", msg.NodeID)
		n.handleOK(msg.NodeID)
	case protocol.Coordinator:
		n.log.Info("recv COORDINATOR", "from", msg.NodeID)
		n.handleCoordinator(msg.NodeID)
	}
}

func (n *Node) getHigherPeers() []string {
	var higher []string
	for peerID, addr := range n.Peers {
		if peerID > n.ID {
			higher = append(higher, addr)
		}
	}
	return higher
}

func (n *Node) broadcast(msg protocol.Message) {
	for _, addr := range n.Peers {
		peerAddr := addr
		go func() {
			if err := transport.SendMessage(peerAddr, msg); err != nil {
				n.log.Warn("broadcast failed", "peer", peerAddr, "error", err)
			}
		}()
	}
	if n.ProxyAddr != "" {
		go func() {
			if err := transport.SendMessage(n.ProxyAddr, msg); err != nil {
				n.log.Debug("proxy send failed", "proxy", n.ProxyAddr, "error", err)
			}
		}()
	}
}

func (n *Node) scheduleStartupElection() {
	delay := time.Duration(n.ID) * config.StartupDelay
	time.Sleep(delay)

	select {
	case n.events <- Event{Type: StartupElection}:
	case <-n.ctx.Done():
	}
}

func (n *Node) handleStartupElection() {
	n.mu.Lock()
	hasCoordinator := n.CoordinatorID != 0
	n.mu.Unlock()

	if !hasCoordinator {
		n.log.Info("no coordinator found on startup, starting election")
		n.startElection()
	}
}

func (n *Node) startHTTPServer() {
	if n.httpAddr == "" {
		return
	}

	n.mu.Lock()
	if n.httpServer != nil {
		n.mu.Unlock()
		return
	}
	n.mu.Unlock()

	mux := http.NewServeMux()

	if n.UserStore != nil && n.SessionStore != nil {
		mux.HandleFunc("/api/login", auth.LoginHandler(n.UserStore, n.SessionStore))
		mux.HandleFunc("/api/logout", auth.LogoutHandler(n.SessionStore))
		protected := auth.AuthMiddleware(n.SessionStore)
		mux.Handle("/api/me", protected(http.HandlerFunc(auth.MeHandler())))
	}

	mux.HandleFunc("/", n.frontendHandler)

	server := &http.Server{
		Addr:    n.httpAddr,
		Handler: mux,
	}

	n.mu.Lock()
	n.httpServer = server
	n.mu.Unlock()

	go func() {
		n.log.Info("frontend HTTP server started", "addr", n.httpAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			n.log.Warn("frontend HTTP server closed", "error", err)
		}
	}()
}

func (n *Node) stopHTTPServer() {
	n.mu.Lock()
	server := n.httpServer
	n.httpServer = nil
	n.mu.Unlock()

	if server != nil {
		n.log.Info("stopping frontend HTTP server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			n.log.Warn("frontend HTTP server shutdown error", "error", err)
		}
	}
}

func (n *Node) frontendHandler(w http.ResponseWriter, r *http.Request) {
	if n.frontendDir == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<html>
<head><title>Red Hospitalaria Distribuida</title></head>
<body>
<h1>Red Hospitalaria Distribuida</h1>
<p>Coordinador: Nodo %d</p>
<p>IP: %s</p>
</body>
</html>`, n.ID, n.IP)
		return
	}

	path := filepath.Join(n.frontendDir, r.URL.Path)
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		http.ServeFile(w, r, filepath.Join(n.frontendDir, "index.html"))
		return
	}
	http.ServeFile(w, r, path)
}

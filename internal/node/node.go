package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/api"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/auth"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/data"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/lock"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/transport"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/pkg/config"
)

func stateString(s NodeState) string {
	switch s {
	case Follower:
		return "FOLLOWER"
	case Discovering:
		return "DISCOVERING"
	case Candidate:
		return "CANDIDATE"
	case Coordinator:
		return "COORDINATOR"
	default:
		return "UNKNOWN"
	}
}

type NodeState int

const (
	Follower    NodeState = iota
	Discovering
	Candidate
	Coordinator
)

type EventType int

const (
	MsgReceived EventType = iota
	ElectionResponseTimeout
	StartupElection
	StartupDiscovery
	DiscoveryTimeout
	HeartbeatTick
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

	watchdogTicker      *time.Ticker
	discoveryTimer      *time.Timer
	electionTimer       *time.Timer
	electionSeq         int64
	currentElectionID   int64

	httpServer     *http.Server
	httpAddr       string // empty = do not start HTTP server
	frontendDir    string // path to built SPA (frontend/dist)

	UserStore       *data.UserStore
	SessionStore    *auth.SessionStore
	PacienteStore   *data.GenericStore[*data.Paciente]
	DonanteStore    *data.GenericStore[*data.Donante]
	OrganoStore     *data.GenericStore[*data.Organo]
	TrasplanteStore *data.GenericStore[*data.Trasplante]

	logBuffer *LogBuffer
	logQueue  chan protocol.LogEntry

	BullyEnabled bool
	LockManager  *lock.LockManager

	timeOffset     time.Duration
	timeSyncMu     sync.Mutex

	currentEpoch   int
	lastHeartbeatAt time.Time
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
		Port:            port,
		Peers:           peers,
		ProxyAddr:       config.ProxyAddr,
		State:           Follower,
		ctx:             ctx,
		cancel:          cancel,
		events:          make(chan Event, 64),
		log:             slog.With("node_id", id, "ip", ip),
		logBuffer:       NewLogBuffer(1000),
		logQueue:        make(chan protocol.LogEntry, 256),
		BullyEnabled:    true,
		lastHeartbeatAt: time.Now(),
	}
}

func (n *Node) SetLogger(logger *slog.Logger) {
	n.log = logger
}

func (n *Node) SetHTTPAddr(addr string) {
	n.httpAddr = addr
}

func (n *Node) SetFrontendDir(dir string) {
	n.frontendDir = dir
}

func (n *Node) IsBullyEnabled() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.BullyEnabled
}

func (n *Node) SetBullyEnabled(enabled bool) {
	n.mu.Lock()
	n.BullyEnabled = enabled
	n.mu.Unlock()
	n.log.Info("bully algorithm toggled", "enabled", enabled, "category", "bully")
	if enabled {
		// Force an immediate election check
		go func() {
			select {
			case n.events <- Event{Type: StartupElection}:
			case <-n.ctx.Done():
			}
		}()
	}
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

	n.watchdogTicker = time.NewTicker(config.HeartbeatTimeout / 2)
	go n.eventLoop()
	go n.logWorker()
	go n.scheduleStartupDiscovery()

	n.startHTTPServer()
	n.startPeriodicTimeSync()
}

func (n *Node) Stop() {
	n.log.Info("stopping node")
	n.cancel()
	if n.watchdogTicker != nil {
		n.watchdogTicker.Stop()
	}
	if n.discoveryTimer != nil {
		n.discoveryTimer.Stop()
	}
	if n.electionTimer != nil {
		n.electionTimer.Stop()
	}
	n.stopHTTPServer()
	if n.LockManager != nil {
		n.LockManager.Stop()
	}
	if n.server != nil {
		n.server.Stop()
	}
}

func (n *Node) Now() time.Time {
	n.timeSyncMu.Lock()
	offset := n.timeOffset
	n.timeSyncMu.Unlock()
	return time.Now().Add(offset)
}

func (n *Node) syncWithCoordinator() {
	n.mu.Lock()
	coordID := n.CoordinatorID
	n.mu.Unlock()

	if coordID == 0 || coordID == n.ID {
		return
	}

	coordIP, ok := config.IDToIP[coordID]
	if !ok {
		return
	}

	url := fmt.Sprintf("http://%s:%s/api/internal/time", coordIP, config.FrontendPort)
	client := &http.Client{Timeout: 5 * time.Second}

	T1 := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		n.log.Warn("time sync request failed", "error", err, "coordinator", coordIP)
		return
	}
	defer resp.Body.Close()
	T2 := time.Now()

	var body struct {
		ServerTime int64 `json:"server_time"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		n.log.Warn("time sync decode failed", "error", err)
		return
	}

	serverAt := time.Unix(0, body.ServerTime)
	oneWay := T2.Sub(T1) / 2
	serverTimeAtT2 := serverAt.Add(oneWay)
	offset := serverTimeAtT2.Sub(T2)

	n.timeSyncMu.Lock()
	n.timeOffset = offset
	n.timeSyncMu.Unlock()

	n.log.Info("time synchronized", "offset_ns", offset, "rtt", T2.Sub(T1))
}

func (n *Node) timeSyncHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{
		"server_time": now.UnixNano(),
	})
}

func (n *Node) startPeriodicTimeSync() {
	n.log.Info("starting periodic time sync", "interval", config.TimeSyncInterval, "category", "system")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				n.log.Error("TIMESYNC_PANIC", "recover", r, "stack", string(debug.Stack()), "category", "system")
			}
		}()
		ticker := time.NewTicker(config.TimeSyncInterval)
		defer ticker.Stop()
		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				n.syncWithCoordinator()
			}
		}
	}()
}

func (n *Node) onMessage(msg protocol.Message, addr net.Addr) {
	select {
	case n.events <- Event{Type: MsgReceived, Msg: &msg, FromAddr: addr.String()}:
	case <-n.ctx.Done():
	}
}

func (n *Node) eventLoop() {
	statusTicker := time.NewTicker(10 * time.Second)
	defer statusTicker.Stop()
	for {
		if !n.eventLoopIteration(statusTicker) {
			return
		}
	}
}

func (n *Node) eventLoopIteration(st *time.Ticker) bool {
	defer func() {
		if r := recover(); r != nil {
			n.log.Error("EVENT_LOOP_PANIC", "recover", r, "stack", string(debug.Stack()), "category", "system")
		}
	}()
	select {
	case <-n.ctx.Done():
		return false
	case <-st.C:
		n.mu.Lock()
		s := n.State
		c := n.CoordinatorID
		e := n.currentEpoch
		h := time.Since(n.lastHeartbeatAt).String()
		g := n.gotOK
		n.mu.Unlock()
		n.log.Warn("STATUS", "state", stateString(s), "coord", c, "epoch", e, "last_hb_ago", h, "gotOK", g, "category", "system")
	case <-n.watchdogTicker.C:
		n.checkHeartbeatWatch()
	case evt := <-n.events:
		n.handleEvent(evt)
	}
	return true
}

func (n *Node) checkHeartbeatWatch() {
	n.mu.Lock()
	if n.State != Follower {
		n.mu.Unlock()
		return
	}
	if n.Now().Sub(n.lastHeartbeatAt) < config.HeartbeatTimeout {
		n.mu.Unlock()
		return
	}
	bullyEnabled := n.BullyEnabled
	n.CoordinatorID = 0
	n.mu.Unlock()

	n.log.Warn("heartbeat timeout — coordinator may be down", "category", "bully")
	if bullyEnabled {
		n.log.Info("starting election after heartbeat timeout", "category", "bully")
		n.beginElection()
	} else {
		n.log.Warn("bully algorithm disabled — not starting election", "category", "bully")
	}
}

func (n *Node) handleEvent(evt Event) {
	switch evt.Type {
	case MsgReceived:
		if evt.Msg != nil {
			n.dispatchMessage(*evt.Msg, evt.FromAddr)
		}
	case ElectionResponseTimeout:
		n.handleElectionTimeout()
	case StartupElection:
		n.handleStartupElection()
	case StartupDiscovery:
		n.handleStartupDiscovery()
	case DiscoveryTimeout:
		n.handleDiscoveryTimeout()
	case HeartbeatTick:
		n.handleHeartbeatTick()
	}
}

func (n *Node) dispatchMessage(msg protocol.Message, fromAddr string) {
	switch msg.Type {
	case protocol.Heartbeat:
		n.log.Debug("recv HEARTBEAT", "category", "heartbeat", "from", msg.NodeID, "coord", msg.CoordinatorID, "epoch", msg.Epoch)
		n.handleHeartbeatMsg(msg.NodeID, msg.CoordinatorID, msg.Epoch)
	case protocol.Election:
		n.log.Debug("recv ELECTION", "category", "bully", "from", msg.NodeID, "epoch", msg.Epoch, "election_id", msg.ElectionID)
		n.handleElectionMsg(msg.NodeID, fromAddr, msg.Epoch, msg.ElectionID)
	case protocol.OK:
		n.log.Debug("recv OK", "category", "bully", "from", msg.NodeID, "epoch", msg.Epoch, "election_id", msg.ElectionID)
		n.handleOKMsg(msg.NodeID, msg.Epoch, msg.ElectionID)
	case protocol.Coordinator:
		n.log.Info("recv COORDINATOR", "category", "bully", "from", msg.NodeID, "epoch", msg.Epoch)
		n.handleCoordinatorMsg(msg.NodeID, msg.Epoch)
	case protocol.CoordinatorQuery:
		n.log.Debug("recv COORDINATOR_QUERY", "category", "bully", "from", msg.NodeID)
		n.handleCoordinatorQuery(msg.NodeID, fromAddr, msg.Epoch)
	case protocol.CoordinatorAck:
		n.log.Debug("recv COORDINATOR_ACK", "category", "bully", "from", msg.NodeID, "coord", msg.CoordinatorID, "epoch", msg.Epoch)
		n.handleCoordinatorAck(msg.NodeID, msg.CoordinatorID, msg.Epoch)
	case protocol.LogEvent:
		if n.State == Coordinator && msg.LogData != nil {
			n.logBuffer.Add(*msg.LogData)
		}
	case protocol.SyncEvent:
		if msg.SyncPayload != nil {
			n.handleSyncEvent(msg.NodeID, msg.SyncPayload, msg.Epoch)
		}
	}
}

func (n *Node) safeSend(addr string, msg protocol.Message) {
	defer func() {
		if r := recover(); r != nil {
			n.log.Error("SEND_PANIC", "peer", addr, "recover", r, "stack", string(debug.Stack()))
		}
	}()
	if err := transport.SendMessage(addr, msg); err != nil {
		n.log.Warn("send failed", "peer", addr, "error", err)
	}
}

func (n *Node) broadcast(msg protocol.Message) {
	for _, addr := range n.Peers {
		peerAddr := addr
		go n.safeSend(peerAddr, msg)
	}
	if n.ProxyAddr != "" {
		go n.safeSend(n.ProxyAddr, msg)
	}
}

func (n *Node) scheduleStartupDiscovery() {
	delay := time.Duration(rand.Int63n(int64(config.StartupDelayMax)))
	time.Sleep(delay)

	select {
	case n.events <- Event{Type: StartupDiscovery}:
	case <-n.ctx.Done():
	}
}

// randomTimeout returns base + random[0, jitter).
func randomTimeout(base, jitter time.Duration) time.Duration {
	if jitter <= 0 || base <= 0 {
		return base
	}
	return base + time.Duration(rand.Int63n(int64(jitter)))
}

func (n *Node) captureLog(entry protocol.LogEntry) {
	if entry.Category == "heartbeat" {
		return
	}

	n.mu.Lock()
	isCoord := n.State == Coordinator
	coordID := n.CoordinatorID
	n.mu.Unlock()

	if isCoord {
		n.logBuffer.Add(entry)
	} else if coordID > 0 && coordID != n.ID {
		select {
		case n.logQueue <- entry:
		default:
		}
	}
}

func (n *Node) logWorker() {
	defer func() {
		if r := recover(); r != nil {
			n.log.Error("LOGWORKER_PANIC", "recover", r, "stack", string(debug.Stack()), "category", "system")
		}
	}()
	for {
		select {
		case <-n.ctx.Done():
			return
		case entry := <-n.logQueue:
			n.sendLog(entry)
		}
	}
}

func (n *Node) sendLog(entry protocol.LogEntry) {
	n.mu.Lock()
	addr, ok := n.Peers[n.CoordinatorID]
	n.mu.Unlock()
	if !ok {
		return
	}
	msg := protocol.Message{
		Type:      protocol.LogEvent,
		NodeID:    n.ID,
		Timestamp: entry.Timestamp,
		LogData:   &entry,
	}
	_ = transport.SendMessage(addr, msg)
}

func (n *Node) getLogsHandler(w http.ResponseWriter, r *http.Request) {
	filterNodeID := r.URL.Query().Get("node_id")
	filterLevel := r.URL.Query().Get("level")
	filterCategory := r.URL.Query().Get("category")

	logs := n.logBuffer.GetAll()

	var filtered []protocol.LogEntry
	for _, log := range logs {
		if filterNodeID != "" && fmt.Sprint(log.NodeID) != filterNodeID {
			continue
		}
		if filterLevel != "" && log.Level != filterLevel {
			continue
		}
		if filterCategory != "" && log.Category != filterCategory {
			continue
		}
		filtered = append(filtered, log)
	}
	if filtered == nil {
		filtered = []protocol.LogEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
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
	forwardMux := n.apiForwardMiddleware(mux)
	loggedMux := n.apiLogMiddleware(forwardMux)

	if n.UserStore != nil && n.SessionStore != nil {
		mux.HandleFunc("/api/login", auth.LoginHandler(n.UserStore, n.SessionStore))
		mux.HandleFunc("/api/logout", auth.LogoutHandler(n.SessionStore))
		protected := auth.AuthMiddleware(n.SessionStore)
		adminProtected := adminMiddleware(protected)
		mux.Handle("/api/me", protected(http.HandlerFunc(auth.MeHandler())))

		if n.PacienteStore != nil {
			pacienteAPI := &api.EntityAPI[*data.Paciente]{Store: n.PacienteStore, Model: "paciente", OnWrite: n.syncCallback()}
			mux.Handle("GET /api/pacientes", protected(http.HandlerFunc(pacienteAPI.List)))
			mux.Handle("GET /api/pacientes/{id}", protected(http.HandlerFunc(pacienteAPI.Get)))
			mux.Handle("POST /api/pacientes", protected(http.HandlerFunc(pacienteAPI.Create)))
			mux.Handle("PUT /api/pacientes/{id}", protected(http.HandlerFunc(pacienteAPI.Update)))
			mux.Handle("DELETE /api/pacientes/{id}", protected(http.HandlerFunc(pacienteAPI.Delete)))
		}

		if n.DonanteStore != nil {
			donanteAPI := &api.EntityAPI[*data.Donante]{Store: n.DonanteStore, Model: "donante", OnWrite: n.syncCallback()}
			mux.Handle("GET /api/donantes", protected(http.HandlerFunc(donanteAPI.List)))
			mux.Handle("GET /api/donantes/{id}", protected(http.HandlerFunc(donanteAPI.Get)))
			mux.Handle("POST /api/donantes", protected(http.HandlerFunc(donanteAPI.Create)))
			mux.Handle("PUT /api/donantes/{id}", protected(http.HandlerFunc(donanteAPI.Update)))
			mux.Handle("DELETE /api/donantes/{id}", protected(http.HandlerFunc(donanteAPI.Delete)))
		}

		if n.OrganoStore != nil {
			organoAPI := &api.EntityAPI[*data.Organo]{Store: n.OrganoStore, Model: "organo", OnWrite: n.syncCallback()}
			mux.Handle("GET /api/organos", protected(http.HandlerFunc(organoAPI.List)))
			mux.Handle("GET /api/organos/{id}", protected(http.HandlerFunc(organoAPI.Get)))
			mux.Handle("POST /api/organos", protected(http.HandlerFunc(organoAPI.Create)))
			mux.Handle("PUT /api/organos/{id}", protected(http.HandlerFunc(organoAPI.Update)))
			mux.Handle("DELETE /api/organos/{id}", protected(http.HandlerFunc(organoAPI.Delete)))

			if n.PacienteStore != nil {
				organoCompatibles := &api.OrganoCompatiblesHandler{
					OrganoStore:   n.OrganoStore,
					PacienteStore: n.PacienteStore,
				}
				mux.Handle("GET /api/organos/compatibles", protected(http.HandlerFunc(organoCompatibles.ListCompatibles)))
			}
		}

		if n.UserStore != nil {
			userAPI := &api.UserAPI{Store: n.UserStore, OnWrite: n.syncCallback()}
			mux.Handle("GET /api/usuarios", adminProtected(http.HandlerFunc(userAPI.List)))
			mux.Handle("GET /api/usuarios/{id}", adminProtected(http.HandlerFunc(userAPI.Get)))
			mux.Handle("POST /api/usuarios", adminProtected(http.HandlerFunc(userAPI.Create)))
			mux.Handle("PUT /api/usuarios/{id}", adminProtected(http.HandlerFunc(userAPI.Update)))
			mux.Handle("DELETE /api/usuarios/{id}", adminProtected(http.HandlerFunc(userAPI.Delete)))
		}

		mux.Handle("GET /api/admin/logs", adminProtected(http.HandlerFunc(n.getLogsHandler)))

		bullyHandler := &api.BullyHandler{
			IsEnabled:  n.IsBullyEnabled,
			SetEnabled: n.SetBullyEnabled,
		}
		mux.Handle("GET /api/admin/bully", adminProtected(http.HandlerFunc(bullyHandler.GetStatus)))
		mux.Handle("POST /api/admin/bully", adminProtected(http.HandlerFunc(bullyHandler.SetStatus)))

		if n.TrasplanteStore != nil {
			trasplanteAPI := &api.EntityAPI[*data.Trasplante]{Store: n.TrasplanteStore, Model: "trasplante", OnWrite: n.syncCallback()}
			mux.Handle("GET /api/trasplantes", protected(http.HandlerFunc(trasplanteAPI.List)))
			mux.Handle("GET /api/trasplantes/{id}", protected(http.HandlerFunc(trasplanteAPI.Get)))
			mux.Handle("POST /api/trasplantes", protected(http.HandlerFunc(trasplanteAPI.Create)))
			mux.Handle("PUT /api/trasplantes/{id}", protected(http.HandlerFunc(trasplanteAPI.Update)))
			mux.Handle("DELETE /api/trasplantes/{id}", protected(http.HandlerFunc(trasplanteAPI.Delete)))
		}

		if n.LockManager != nil {
			lockHandler := &lock.Handler{Manager: n.LockManager}
			mux.Handle("POST /api/lock", protected(http.HandlerFunc(lockHandler.Acquire)))
			mux.Handle("POST /api/unlock", protected(http.HandlerFunc(lockHandler.Release)))
			mux.Handle("GET /api/locks", protected(http.HandlerFunc(lockHandler.List)))
		}
	}

	mux.HandleFunc("/api/internal/time", n.timeSyncHandler)

	mux.HandleFunc("/", n.frontendHandler)

	server := &http.Server{
		Addr:    n.httpAddr,
		Handler: loggedMux,
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

func adminMiddleware(authMW func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return authMW(auth.AdminMiddleware(next))
	}
}

func (n *Node) apiLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			rw := &responseWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)
			n.captureLog(protocol.LogEntry{
				NodeID:    n.ID,
				Level:     "INFO",
				Category:  "api",
				Message:   r.Method + " " + r.URL.Path + " -> " + strconv.Itoa(rw.status),
				Timestamp: time.Now().Unix(),
			})
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (n *Node) apiForwardMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		n.mu.Lock()
		isCoord := n.State == Coordinator
		coordID := n.CoordinatorID
		n.mu.Unlock()

		if !isCoord && coordID > 0 {
			n.forwardToCoordinator(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (n *Node) forwardToCoordinator(w http.ResponseWriter, r *http.Request) {
	n.mu.Lock()
	coordID := n.CoordinatorID
	n.mu.Unlock()

	coordIP, ok := config.IDToIP[coordID]
	if !ok {
		http.Error(w, `{"error":"coordinador no disponible"}`, http.StatusServiceUnavailable)
		return
	}

	coordAddr := net.JoinHostPort(coordIP, config.FrontendPort)
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = coordAddr
			req.Host = coordAddr
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			n.log.Warn("forward to coordinator failed", "error", err, "coordinator", coordAddr)
			http.Error(w, `{"error":"coordinador no responde"}`, http.StatusBadGateway)
		},
	}
	proxy.ServeHTTP(w, r)
}

func (n *Node) syncCallback() func(action, model string, id int, data json.RawMessage) {
	return func(action, model string, id int, data json.RawMessage) {
		n.mu.Lock()
		coordID := n.CoordinatorID
		isCoord := n.State == Coordinator
		epoch := n.currentEpoch
		n.mu.Unlock()
		if !isCoord || coordID != n.ID {
			return
		}
		n.broadcastSync(&protocol.SyncPayload{
			Model:  model,
			Action: action,
			ID:     id,
			Data:   data,
		}, epoch)
	}
}

func (n *Node) broadcastSync(payload *protocol.SyncPayload, epoch int) {
	msg := protocol.Message{
		Type:        protocol.SyncEvent,
		NodeID:      n.ID,
		Timestamp:   time.Now().Unix(),
		Epoch:       epoch,
		SyncPayload: payload,
	}
	for _, addr := range n.Peers {
		peerAddr := addr
		go n.safeSend(peerAddr, msg)
	}
}

func (n *Node) handleSyncEvent(fromID int, payload *protocol.SyncPayload, msgEpoch int) {
	n.mu.Lock()
	if msgEpoch > 0 && msgEpoch < n.currentEpoch {
		n.mu.Unlock()
		n.log.Warn("rejecting stale SYNC_EVENT", "from", fromID, "msg_epoch", msgEpoch, "current_epoch", n.currentEpoch)
		return
	}
	if msgEpoch > n.currentEpoch {
		n.currentEpoch = msgEpoch
	}
	n.mu.Unlock()

	n.log.Debug("received sync event", "from", fromID, "action", payload.Action, "model", payload.Model, "id", payload.ID)

	switch payload.Model {
	case "paciente":
		n.syncGeneric(n.PacienteStore, payload)
	case "donante":
		n.syncGeneric(n.DonanteStore, payload)
	case "organo":
		n.syncGeneric(n.OrganoStore, payload)
	case "trasplante":
		n.syncGeneric(n.TrasplanteStore, payload)
	case "usuario":
		n.syncUsers(payload)
	}
}

func (n *Node) syncGeneric(store any, payload *protocol.SyncPayload) {
	if store == nil {
		return
	}
	switch s := store.(type) {
	case *data.GenericStore[*data.Paciente]:
		syncStore(s, payload, n.log)
	case *data.GenericStore[*data.Donante]:
		syncStore(s, payload, n.log)
	case *data.GenericStore[*data.Organo]:
		syncStore(s, payload, n.log)
	case *data.GenericStore[*data.Trasplante]:
		syncStore(s, payload, n.log)
	}
}

func syncStore[T data.Entity](s *data.GenericStore[T], payload *protocol.SyncPayload, log *slog.Logger) {
	switch payload.Action {
	case "create":
		if len(payload.Data) == 0 {
			return
		}
		var item T
		if err := json.Unmarshal(payload.Data, &item); err != nil {
			log.Warn("sync unmarshal failed", "model", payload.Model, "error", err)
			return
		}
		if err := s.SyncCreate(item); err != nil {
			log.Warn("sync create failed", "model", payload.Model, "error", err)
		}
	case "update":
		if len(payload.Data) == 0 {
			return
		}
		var item T
		if err := json.Unmarshal(payload.Data, &item); err != nil {
			log.Warn("sync unmarshal failed", "model", payload.Model, "error", err)
			return
		}
		if err := s.SyncUpdate(item); err != nil {
			log.Warn("sync update failed", "model", payload.Model, "error", err)
		}
	case "delete":
		if err := s.Delete(payload.ID); err != nil {
			log.Warn("sync delete failed", "model", payload.Model, "error", err)
		}
	}
}

func (n *Node) syncUsers(payload *protocol.SyncPayload) {
	if n.UserStore == nil {
		return
	}
	switch payload.Action {
	case "create":
		if len(payload.Data) == 0 {
			return
		}
		var user data.User
		if err := json.Unmarshal(payload.Data, &user); err != nil {
			n.log.Warn("sync user unmarshal failed", "error", err)
			return
		}
		if err := n.UserStore.SyncCreate(&user); err != nil {
			n.log.Warn("sync user create failed", "error", err)
		}
	case "update":
		if len(payload.Data) == 0 {
			return
		}
		var user data.User
		if err := json.Unmarshal(payload.Data, &user); err != nil {
			n.log.Warn("sync user unmarshal failed", "error", err)
			return
		}
		if err := n.UserStore.SyncUpdate(&user); err != nil {
			n.log.Warn("sync user update failed", "error", err)
		}
	case "delete":
		if err := n.UserStore.Delete(payload.ID); err != nil {
			n.log.Warn("sync user delete failed", "error", err)
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

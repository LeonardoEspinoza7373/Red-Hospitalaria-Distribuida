package proxy

import (
	"context"
	"fmt"
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
	wasStale        bool

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

	wasChange := id != p.coordinatorID

	if wasChange {
		p.coordinatorID = id
		if addr, ok := p.nodeAddrs[id]; ok {
			p.coordinatorAddr = addr
		}
	}

	if p.wasStale {
		p.wasStale = false
		p.log.Warn(">>> COORDINADOR RECUPERADO <<<", "coordinator_id", id, "addr", p.coordinatorAddr)
	} else if wasChange {
		p.log.Info("coordinator changed", "old", p.coordinatorID, "new", id)
	}
	p.lastHeartbeat = time.Now()
}

func (p *Proxy) getCoordinatorAddr() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.coordinatorAddr
}

func (p *Proxy) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/status" {
		p.serveStatus(w)
		return
	}

	addr := p.getCoordinatorAddr()
	if addr == "" {
		p.log.Warn(">>> SERVIDOR NO DISPONIBLE - coordinador no encontrado <<<")
		p.serveUnavailable(w)
		return
	}
	r.URL.Host = addr
	r.URL.Scheme = "http"
	p.rp.ServeHTTP(w, r)
}

func (p *Proxy) serveStatus(w http.ResponseWriter) {
	p.mu.Lock()
	id := p.coordinatorID
	addr := p.coordinatorAddr
	since := time.Since(p.lastHeartbeat)
	ws := p.wasStale
	p.mu.Unlock()

	statusClass := "ok"
	statusText := "OPERATIVO"
	coordText := fmt.Sprintf("Nodo %d (%s)", id, addr)
	if id == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		statusClass = "err"
		statusText = "NO DISPONIBLE"
		coordText = "ninguno"
	}

	recovery := "no"
	if ws {
		recovery = "sí (recuperado de caída)"
	}

	fmt.Fprintf(w, `<html>
<head><title>Red Hospitalaria - Proxy</title>
<style>
body{font-family:monospace;background:#0f172a;color:#e2e8f0;padding:2rem}
h1{color:#38bdf8;font-size:1.3rem}
.ok{color:#4ade80}.warn{color:#facc15}.err{color:#f87171}
pre{background:#1e293b;padding:1rem;border-radius:8px}
</style>
</head><body>
<h1>Proxy - Red Hospitalaria Distribuida</h1>
<pre>
Estado:         <span class="%s">%s</span>
Coordinador:    %s
Último latido:  %v atrás
Recuperación:   %v
</pre>
<p><a href="/">Volver al inicio</a></p>
</body></html>`, statusClass, statusText, coordText, since.Round(time.Second), recovery)
}

func (p *Proxy) serveUnavailable(w http.ResponseWriter) {
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte(`<html>
<head><title>Red Hospitalaria - No Disponible</title>
<style>
body{font-family:monospace;background:#0f172a;color:#e2e8f0;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;flex-direction:column;text-align:center;padding:2rem}
h1{color:#f87171;font-size:2rem;margin-bottom:0.5rem}
p{color:#94a3b8;max-width:400px;line-height:1.6}
code{color:#38bdf8}
.refresh{color:#64748b;font-size:0.85rem;margin-top:2rem}
</style>
</head><body>
<h1>Sistema No Disponible</h1>
<p>El coordinador actual ha fallado. Los nodos están ejecutando una nueva elección. Por favor espere...</p>
<p><code>Intente nuevamente en unos segundos</code></p>
<p class="refresh">La página se recargará automáticamente <span id="secs">15</span>s</p>
<script>
let s=15;setInterval(()=>{s--;document.getElementById('secs').textContent=s;if(s<=0)location.reload()},1000)
</script>
</body></html>`))
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
		p.log.Warn("🔥🔥 COORDINADOR CAÍDO - Iniciando elección en los nodos... 🔥🔥")
				p.clearCoordinator()
			}
		}
	}
}

func (p *Proxy) clearCoordinator() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.wasStale = true
	p.coordinatorID = 0
	p.coordinatorAddr = ""
}

func (p *Proxy) CurrentCoordinator() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.coordinatorID
}

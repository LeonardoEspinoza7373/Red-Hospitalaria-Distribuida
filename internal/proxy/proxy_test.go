package proxy

import (
	"log/slog"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/transport"
)

func TestProxyDiscoversCoordinatorViaTCP(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))

	basePort := 9590

	p := NewWithPorts(strconv.Itoa(basePort), strconv.Itoa(basePort+100))
	if err := p.Start(); err != nil {
		t.Fatalf("failed to start proxy: %v", err)
	}
	defer p.Stop()

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(basePort))
	msg := protocol.NewCoordinatorMessage(4)
	if err := transport.SendMessage(addr, msg); err != nil {
		t.Fatalf("failed to send COORDINATOR: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if got := p.CurrentCoordinator(); got != 4 {
		t.Errorf("expected coordinator 4, got %d", got)
	}

	msg2 := protocol.NewCoordinatorMessage(3)
	if err := transport.SendMessage(addr, msg2); err != nil {
		t.Fatalf("failed to send COORDINATOR: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if got := p.CurrentCoordinator(); got != 3 {
		t.Errorf("expected coordinator 3 after change, got %d", got)
	}
}

func TestProxyDiscoversViaHeartbeat(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))

	basePort := 9560

	p := NewWithPorts(strconv.Itoa(basePort), strconv.Itoa(basePort+100))
	if err := p.Start(); err != nil {
		t.Fatalf("failed to start proxy: %v", err)
	}
	defer p.Stop()

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(basePort))
	msg := protocol.NewHeartbeatMessage(4, 4)
	if err := transport.SendMessage(addr, msg); err != nil {
		t.Fatalf("failed to send HEARTBEAT: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if got := p.CurrentCoordinator(); got != 4 {
		t.Errorf("expected coordinator 4 via heartbeat, got %d", got)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

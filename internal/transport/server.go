package transport

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"runtime/debug"
	"sync"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
)

type MessageHandler func(protocol.Message, net.Addr)

type Server struct {
	addr    string
	handler MessageHandler
	ln      net.Listener
	wg      sync.WaitGroup
	closed  chan struct{}
}

func NewServer(addr string, handler MessageHandler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
		closed:  make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("tcp listen on %s: %w", s.addr, err)
	}
	s.ln = ln
	slog.Info("TCP server listening", "addr", s.addr)

	s.wg.Add(1)
	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-s.closed:
				return
			default:
				slog.Error("accept error", "error", err)
				continue
			}
		}
		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("TCP_HANDLER_PANIC", "recover", r, "stack", string(debug.Stack()))
		}
	}()

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		msg, err := protocol.Decode(scanner.Bytes())
		if err != nil {
			slog.Error("decode error", "error", err)
			continue
		}
		s.handler(msg, conn.RemoteAddr())
	}

	if err := scanner.Err(); err != nil {
		slog.Error("scan error", "error", err)
	}
}

func (s *Server) Stop() {
	select {
	case <-s.closed:
	default:
		close(s.closed)
	}
	if s.ln != nil {
		s.ln.Close()
	}
	s.wg.Wait()
}

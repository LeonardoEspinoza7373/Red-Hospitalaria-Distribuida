package node

import (
	"context"
	"log/slog"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
)

type CaptureHandler struct {
	inner slog.Handler
	node  *Node
}

func NewCaptureHandler(inner slog.Handler, node *Node) *CaptureHandler {
	return &CaptureHandler{inner: inner, node: node}
}

func (h *CaptureHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *CaptureHandler) Handle(ctx context.Context, r slog.Record) error {
	entry := protocol.LogEntry{
		NodeID:    h.node.ID,
		Level:     r.Level.String(),
		Message:   r.Message,
		Timestamp: r.Time.Unix(),
	}
	h.node.captureLog(entry)
	return h.inner.Handle(ctx, r)
}

func (h *CaptureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CaptureHandler{
		inner: h.inner.WithAttrs(attrs),
		node:  h.node,
	}
}

func (h *CaptureHandler) WithGroup(name string) slog.Handler {
	return &CaptureHandler{
		inner: h.inner.WithGroup(name),
		node:  h.node,
	}
}

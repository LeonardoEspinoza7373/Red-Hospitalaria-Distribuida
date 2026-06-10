package node

import (
	"sync"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
)

type LogBuffer struct {
	mu      sync.RWMutex
	entries []protocol.LogEntry
	maxSize int
}

func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		entries: make([]protocol.LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (b *LogBuffer) Add(entry protocol.LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.entries) >= b.maxSize {
		b.entries = b.entries[1:]
	}
	b.entries = append(b.entries, entry)
}

func (b *LogBuffer) GetAll() []protocol.LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]protocol.LogEntry, len(b.entries))
	copy(result, b.entries)
	return result
}

func (b *LogBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = b.entries[:0]
}

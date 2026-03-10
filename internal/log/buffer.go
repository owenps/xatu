package log

import (
	"fmt"
	"sync"

	"github.com/owen/xatu/internal/aws"
)

// Buffer is a thread-safe ring buffer for log entries with deduplication.
type Buffer struct {
	mu      sync.RWMutex
	entries []aws.LogEntry
	seen    map[string]struct{}
	maxSize int
}

func NewBuffer(maxSize int) *Buffer {
	return &Buffer{
		entries: make([]aws.LogEntry, 0, maxSize),
		seen:    make(map[string]struct{}, maxSize),
		maxSize: maxSize,
	}
}

func (b *Buffer) Append(entries []aws.LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, e := range entries {
		key := entryKey(e)
		if _, exists := b.seen[key]; exists {
			continue
		}
		b.seen[key] = struct{}{}
		b.entries = append(b.entries, e)
	}

	// Trim to max size, keeping the most recent entries
	if len(b.entries) > b.maxSize {
		excess := len(b.entries) - b.maxSize
		// Remove keys for evicted entries
		for _, e := range b.entries[:excess] {
			delete(b.seen, entryKey(e))
		}
		b.entries = b.entries[excess:]
	}
}

func (b *Buffer) Entries() []aws.LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]aws.LogEntry, len(b.entries))
	copy(out, b.entries)
	return out
}

func (b *Buffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.entries)
}

func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = b.entries[:0]
	b.seen = make(map[string]struct{}, b.maxSize)
}

func entryKey(e aws.LogEntry) string {
	msg := e.Message
	if len(msg) > 100 {
		msg = msg[:100]
	}
	return fmt.Sprintf("%d|%s|%s|%s", e.Timestamp.UnixMilli(), e.LogGroup, e.LogStream, msg)
}

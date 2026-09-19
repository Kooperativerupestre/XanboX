package execution

import (
	"bytes"
	"sync"
)

type synchronizedBuffer struct {
	mu     sync.RWMutex
	buffer bytes.Buffer
}

func (b *synchronizedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buffer.Write(p)
}

func (b *synchronizedBuffer) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.buffer.String()
}

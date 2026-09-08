package live

import "sync"

type broadcastWriter struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

func newBroadcastWriter() *broadcastWriter {
	return &broadcastWriter{
		clients: make(map[chan []byte]struct{}),
	}
}

func (h *broadcastWriter) Write(p []byte) (n int, err error) {
	h.mu.RLock()
	if len(h.clients) == 0 {
		h.mu.RUnlock()
		return len(p), nil
	}

	chunk := make([]byte, len(p))
	copy(chunk, p)

	for clientChan := range h.clients {
		select {
		case clientChan <- chunk:
		default:
		}
	}
	h.mu.RUnlock()

	return len(p), nil
}

func (h *broadcastWriter) AddClient(clientChan chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[clientChan] = struct{}{}
}

func (h *broadcastWriter) RemoveClient(clientChan chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, clientChan)
}

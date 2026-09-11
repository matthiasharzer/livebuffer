package stream

import "sync"

type BroadcastWriter struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

func newBroadcastWriter() *BroadcastWriter {
	return &BroadcastWriter{
		clients: make(map[chan []byte]struct{}),
	}
}

func (h *BroadcastWriter) Write(p []byte) (n int, err error) {
	h.mu.RLock()
	if len(h.clients) == 0 {
		h.mu.RUnlock()
		return len(p), nil
	}

	chunk := make([]byte, len(p))
	copy(chunk, p)

	var laggingClients []chan []byte
	for clientChan := range h.clients {
		select {
		case clientChan <- chunk:
		default:
			laggingClients = append(laggingClients, clientChan)
		}
	}
	h.mu.RUnlock()

	if len(laggingClients) > 0 {
		h.mu.Lock()
		for _, clientChan := range laggingClients {
			if _, ok := h.clients[clientChan]; !ok {
				continue
			}
			delete(h.clients, clientChan)
			close(clientChan)
		}
		h.mu.Unlock()
	}

	return len(p), nil
}

func (h *BroadcastWriter) AddClient(clientChan chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[clientChan] = struct{}{}
}

func (h *BroadcastWriter) RemoveClient(clientChan chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[clientChan]; !ok {
		return
	}
	delete(h.clients, clientChan)
	close(clientChan)
}

func (h *BroadcastWriter) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	for clientChan := range h.clients {
		close(clientChan)
	}
	h.clients = make(map[chan []byte]struct{})
	return nil
}

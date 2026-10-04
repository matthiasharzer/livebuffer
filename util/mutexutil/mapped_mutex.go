package mutexutil

import "sync"

type MappedMutex struct {
	mu       sync.Mutex
	mutexMap map[string]*mutexEntry
}

type mutexEntry struct {
	refCount int
	mu       *sync.RWMutex
}

func NewMappedMutex() *MappedMutex {
	return &MappedMutex{
		mutexMap: make(map[string]*mutexEntry),
	}
}

func (m *MappedMutex) GetMutex(key string) (*sync.RWMutex, func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.mutexMap[key]
	if !exists {
		entry = &mutexEntry{
			refCount: 0,
			mu:       &sync.RWMutex{},
		}
		m.mutexMap[key] = entry
	}
	entry.refCount++

	cleanupFunc := func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		entry.refCount--
		if entry.refCount == 0 {
			delete(m.mutexMap, key)
		}
	}

	return entry.mu, cleanupFunc
}

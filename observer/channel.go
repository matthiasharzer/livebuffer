package observer

import (
	"maps"
	"sync"
)

type UnsubscribeFunc = func()

type ReadonlyChannel[T any] interface {
	Subscribe(observer func(T)) UnsubscribeFunc
}

type ReadWriteChannel[T any] interface {
	ReadonlyChannel[T]
	Publish(data T)
	Clear()
}

type channel[T any] struct {
	callbacks map[uint64]func(T)
	nextID    uint64
	mu        *sync.RWMutex
}

func NewChannel[T any]() ReadWriteChannel[T] {
	return &channel[T]{
		callbacks: make(map[uint64]func(T)),
		mu:        &sync.RWMutex{},
	}
}

func (o *channel[T]) Subscribe(observer func(T)) UnsubscribeFunc {
	o.mu.Lock()
	defer o.mu.Unlock()

	id := o.nextID
	o.nextID++
	o.callbacks[id] = observer
	return func() {
		o.mu.Lock()
		defer o.mu.Unlock()
		delete(o.callbacks, id)
	}
}

func (o *channel[T]) Publish(data T) {
	o.mu.RLock()
	callbacks := make(map[uint64]func(T), len(o.callbacks))
	maps.Copy(callbacks, o.callbacks)
	o.mu.RUnlock()

	for _, observer := range callbacks {
		observer(data)
	}
}

func (o *channel[T]) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.callbacks = nil
}

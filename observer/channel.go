package observer

import (
	"sync"
)

type UnsubscribeFunc = func()

type Observer[T any] interface {
	Update(data T)
}

type ReadonlyChannel[T any] interface {
	Subscribe(observer Observer[T]) UnsubscribeFunc
	Unsubscribe(observer Observer[T])
}

type ReadWriteChannel[T any] interface {
	ReadonlyChannel[T]
	Publish(data T)
	Clear()
}

type channel[T any] struct {
	callbacks []Observer[T]
	mu        *sync.RWMutex
}

func NewChannel[T any]() ReadWriteChannel[T] {
	return &channel[T]{
		callbacks: make([]Observer[T], 0),
		mu:        &sync.RWMutex{},
	}
}

func (o *channel[T]) Subscribe(observer Observer[T]) UnsubscribeFunc {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.callbacks = append(o.callbacks, observer)
	return func() {
		o.Unsubscribe(observer)
	}
}

func (o *channel[T]) Unsubscribe(observer Observer[T]) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i, obs := range o.callbacks {
		if obs == observer {
			o.callbacks = append(o.callbacks[:i], o.callbacks[i+1:]...)
			break
		}
	}
}

func (o *channel[T]) Publish(data T) {
	o.mu.RLock()
	callbacks := make([]Observer[T], len(o.callbacks))
	copy(callbacks, o.callbacks)
	o.mu.RUnlock()

	for _, observer := range callbacks {
		observer.Update(data)
	}
}

func (o *channel[T]) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.callbacks = nil
}

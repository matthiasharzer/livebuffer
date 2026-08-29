package observer

import "sync"

type UpdateFunc[T any] = func(data T)
type UnsubscribeFunc = func()

type ReadonlyChannel[T any] interface {
	Subscribe(observer UpdateFunc[T]) UnsubscribeFunc
	Unsubscribe(observer UpdateFunc[T])
}

type ReadWriteChannel[T any] interface {
	ReadonlyChannel[T]
	Publish(data T)
	Clear()
}

type channel[T any] struct {
	callbacks []UpdateFunc[T]
	mu        *sync.RWMutex
}

func NewChannel[T any]() ReadWriteChannel[T] {
	return &channel[T]{
		callbacks: make([]UpdateFunc[T], 0),
		mu:        &sync.RWMutex{},
	}
}

func (o *channel[T]) Subscribe(observer UpdateFunc[T]) UnsubscribeFunc {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.callbacks = append(o.callbacks, observer)
	return func() {
		o.Unsubscribe(observer)
	}
}

func (o *channel[T]) Unsubscribe(observer UpdateFunc[T]) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i, obs := range o.callbacks {
		if &obs == &observer {
			o.callbacks = append(o.callbacks[:i], o.callbacks[i+1:]...)
			break
		}
	}
}

func (o *channel[T]) Publish(data T) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, observer := range o.callbacks {
		observer(data)
	}
}

func (o *channel[T]) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.callbacks = nil
}

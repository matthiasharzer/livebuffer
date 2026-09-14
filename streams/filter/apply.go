package filter

import (
	"iter"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
)

func Apply(items iter.Seq2[stream.Manager, error], filterFunc Func) iter.Seq2[stream.Manager, error] {
	if filterFunc == nil {
		return items
	}
	return func(yield func(stream.Manager, error) bool) {
		for item, err := range items {
			if err != nil {
				yield(item, err)
				return
			}
			if !filterFunc(item) {
				continue
			}
			if !yield(item, err) {
				return
			}
		}
	}
}

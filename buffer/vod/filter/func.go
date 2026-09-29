package filter

import (
	"strings"

	"github.com/matthiasharzer/livebuffer/stream"
)

type Func = func(manager stream.Manager) bool

func And(a, b Func) Func {
	return func(manager stream.Manager) bool {
		return a(manager) && b(manager)
	}
}

func ByBroadcasterName(name string) Func {
	return func(manager stream.Manager) bool {
		return strings.EqualFold(manager.Meta().BroadcasterUserName, name)
	}
}

func All() Func {
	return func(manager stream.Manager) bool {
		return true
	}
}

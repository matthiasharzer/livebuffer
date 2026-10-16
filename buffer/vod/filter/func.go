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

func ByAnyBroadcasterNames(names ...string) Func {
	return func(manager stream.Manager) bool {
		for _, name := range names {
			if strings.EqualFold(manager.Meta().BroadcasterUserName, name) {
				return true
			}
		}
		return false
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

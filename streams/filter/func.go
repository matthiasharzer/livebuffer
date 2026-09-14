package filter

import (
	"strings"

	"github.com/matthiasharzer/livebuffer/buffer/stream"
)

type Func = func(manager stream.Manager) bool

func ByBroadcasterName(name string) Func {
	return func(manager stream.Manager) bool {
		return strings.EqualFold(manager.Meta().BroadcasterUserName, name)
	}
}

package list

import (
	"fmt"
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/buffer/vod/filter"
	"github.com/matthiasharzer/livebuffer/stream"
)

type filterBuilder = func(value string, director *buffer.Director) (filter.Func, error)

var filterBuilders = map[string]filterBuilder{
	"username": func(username string, director *buffer.Director) (filter.Func, error) {
		isKnownUser := director.IsObservedBroadcaster(username)
		if !isKnownUser {
			return nil, fmt.Errorf("unknown broadcaster '%s'", username)
		}
		return filter.ByBroadcasterName(username), nil
	},
	"state": func(stateStr string, director *buffer.Director) (filter.Func, error) {
		state := stream.State(stateStr)
		switch state {
		case stream.StateLive, stream.StateArchived:
		default:
			return nil, fmt.Errorf("invalid state '%s'", stateStr)
		}
		return func(manager stream.Manager) bool {
			return director.GetStreamState(manager.StreamID()) == state
		}, nil
	},
}

func getFilter(r *http.Request, w http.ResponseWriter, director *buffer.Director) filter.Func {
	filterFunc := filter.All()
	for key, builder := range filterBuilders {
		value := r.URL.Query().Get(key)
		if value == "" {
			continue
		}
		builtFilter, err := builder(value, director)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil
		}
		filterFunc = filter.And(filterFunc, builtFilter)
	}
	return filterFunc
}

package broadcasters

import (
	"encoding/json"
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/buffer/vod/filter"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/shared"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/stream"
)

func buildResponseStreams(streams []stream.Details) []shared.ResponseStream {
	responseStreams := make([]shared.ResponseStream, 0, len(streams))
	for _, details := range streams {
		responseStreams = append(responseStreams, shared.ResponseStreamFromDetails(details))
	}
	return responseStreams
}

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		broadcasters := director.GetBroadcasterNames()
		streams := director.GetStreams(filter.All())

		streamsByBroadcaster := make(map[string][]stream.Details)
		for _, broadcasterName := range broadcasters {
			streamsByBroadcaster[broadcasterName] = []stream.Details{}
		}
		for s := range streams {
			broadcasterStreams, ok := streamsByBroadcaster[s.BroadcasterUserName]
			if !ok {
				continue
			}
			streamsByBroadcaster[s.BroadcasterUserName] = append(broadcasterStreams, s)
		}

		responseBroadcasters := make([]ResponseBroadcaster, 0, len(broadcasters))
		for _, broadcasterName := range broadcasters {
			streams, ok := streamsByBroadcaster[broadcasterName]
			if !ok {
				continue
			}

			responseBroadcaster := ResponseBroadcaster{
				Username: broadcasterName,
				Streams:  buildResponseStreams(streams),
			}
			responseBroadcasters = append(responseBroadcasters, responseBroadcaster)
		}

		response := Response{
			Broadcasters: responseBroadcasters,
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			logging.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

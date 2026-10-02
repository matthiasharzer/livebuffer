package broadcasters

import (
	"encoding/json"
	"net/http"
	"strings"

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
			streamsByBroadcaster[strings.ToLower(broadcasterName)] = []stream.Details{}
		}
		for s := range streams {
			normalizedUsername := strings.ToLower(s.BroadcasterUserName)

			broadcasterStreams, ok := streamsByBroadcaster[normalizedUsername]
			if !ok {
				continue
			}
			streamsByBroadcaster[normalizedUsername] = append(broadcasterStreams, s)
		}

		responseBroadcasters := make([]ResponseBroadcaster, 0, len(broadcasters))
		for _, broadcasterName := range broadcasters {
			normalizedUsername := strings.ToLower(broadcasterName)

			streams, ok := streamsByBroadcaster[normalizedUsername]
			if !ok {
				logging.Warn("stream disappeared for broadcaster", "broadcaster", broadcasterName)
				continue
			}

			responseBroadcaster := ResponseBroadcaster{
				Username: normalizedUsername,
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

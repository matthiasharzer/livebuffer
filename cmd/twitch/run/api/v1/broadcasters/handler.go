package broadcasters

import (
	"encoding/json"
	"iter"
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/buffer/vod/filter"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/shared"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/stream"
)

func buildResponseStreams(streams iter.Seq[stream.Details]) []shared.ResponseStream {
	var responseStreams []shared.ResponseStream
	for details := range streams {
		responseStreams = append(responseStreams, shared.ResponseStreamFromDetails(details))
	}
	return responseStreams
}

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		broadcasters := director.GetBroadcasterNames()

		var responseBroadcasters []ResponseBroadcaster
		for _, broadcasterName := range broadcasters {
			streams := director.GetStreams(filter.ByBroadcasterName(broadcasterName))

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

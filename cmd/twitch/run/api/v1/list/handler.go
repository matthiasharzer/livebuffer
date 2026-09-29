package list

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/dustin/go-humanize"
	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/shared"
	"github.com/matthiasharzer/livebuffer/logging"
)

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filterFunc := getFilter(r, w, director)
		if filterFunc == nil {
			return
		}
		orderFunc := getOrderFunc(r, w)
		if orderFunc == nil {
			return
		}

		streamsSeq := director.GetStreams(filterFunc)
		sortedStreams := slices.SortedFunc(streamsSeq, orderFunc)

		w.Header().Set("Content-Type", "application/json")
		responseStreams := make([]shared.ResponseStream, 0, len(sortedStreams))
		for _, s := range sortedStreams {
			responseStreams = append(responseStreams, shared.ResponseStream{
				ID:                   s.ID,
				Title:                s.Title,
				Size:                 humanize.Bytes(uint64(s.Size)),
				SizeBytes:            s.Size,
				Duration:             s.Duration.String(),
				DurationMilliseconds: s.Duration.Milliseconds(),
				StartedAt:            s.StartedAt,
				BroadcasterUserName:  s.BroadcasterUserName,
				StreamState:          string(s.StreamState),
			})
		}
		response := Response{
			Streams: responseStreams,
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			logging.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

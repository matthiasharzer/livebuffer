package list

import (
	"encoding/json"
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
)

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streams, err := director.GetStreams()
		if err != nil {
			http.Error(w, "failed to retrieve streams", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		responseStreams := make([]ResponseStream, 0, len(streams))
		for _, stream := range streams {
			responseStreams = append(responseStreams, ResponseStream{
				ID:                  stream.ID,
				Title:               stream.Title,
				Size:                stream.Size,
				Duration:            stream.Duration.String(),
				StartedAt:           stream.StartedAt,
				BroadcasterUserName: stream.BroadcasterUserName,
				StreamState:         string(stream.StreamState),
			})
		}
		response := Response{
			Streams: responseStreams,
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

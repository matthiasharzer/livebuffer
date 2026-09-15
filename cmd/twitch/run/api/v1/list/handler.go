package list

import (
	"encoding/json"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/matthiasharzer/livebuffer/connectorneedrename"
	"github.com/matthiasharzer/livebuffer/logging"
)

func Handler(director *connectorneedrename.Director, username string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allStreams, err := director.Repository.GetStreamsByBroadcaster(username)
		if err != nil {
			logging.Error("failed to retrieve streams", "error", err)
			http.Error(w, "failed to retrieve streams", http.StatusInternalServerError)
			return
		}

		getStreamState := director.GetStreamStateFunc()

		w.Header().Set("Content-Type", "application/json")
		responseStreams := make([]ResponseStream, 0, len(allStreams))
		for _, stream := range allStreams {
			responseStreams = append(responseStreams, ResponseStream{
				ID:                   stream.ID,
				Title:                stream.Title,
				Size:                 humanize.Bytes(uint64(stream.Size)),
				SizeBytes:            stream.Size,
				Duration:             stream.Duration.String(),
				DurationMilliseconds: stream.Duration.Milliseconds(),
				StartedAt:            stream.StartedAt,
				BroadcasterUserName:  stream.BroadcasterUserName,
				StreamState:          string(getStreamState(stream.ID)),
			})
		}
		response := Response{
			Streams: responseStreams,
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logging.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

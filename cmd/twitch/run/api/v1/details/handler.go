package details

import (
	"encoding/json"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/shared"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/stream"
)

func Handler(directory *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamID := r.PathValue("streamID")
		if !stream.IsStreamID(streamID) {
			http.Error(w, "invalid stream_id", http.StatusBadRequest)
			return
		}

		streamDetails, err := directory.GetStream(streamID)
		if err != nil {
			http.Error(w, "failed to retrieve stream info", http.StatusInternalServerError)
			return
		}
		if streamDetails == nil {
			http.Error(w, "stream not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := shared.ResponseStream{
			ID:                   streamDetails.ID,
			Title:                streamDetails.Title,
			Size:                 humanize.Bytes(uint64(streamDetails.Size)),
			SizeBytes:            streamDetails.Size,
			Duration:             streamDetails.Duration.String(),
			DurationMilliseconds: streamDetails.Duration.Milliseconds(),
			StartedAt:            streamDetails.StartedAt,
			BroadcasterUserName:  streamDetails.BroadcasterUserName,
			StreamState:          string(streamDetails.StreamState),
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logging.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

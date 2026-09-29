package details

import (
	"encoding/json"
	"net/http"

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

		response := shared.ResponseStreamFromDetails(*streamDetails)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logging.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

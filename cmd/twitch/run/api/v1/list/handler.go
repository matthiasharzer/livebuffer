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
		response := Response{
			Streams: streams,
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

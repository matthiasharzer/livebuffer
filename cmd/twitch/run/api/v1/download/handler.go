package download

import (
	"io"
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/logging"
)

func Handler(directory *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamID := r.URL.Query().Get("stream_id")
		if streamID == "" {
			http.Error(w, "missing 'stream_id' query parameter", http.StatusBadRequest)
			return
		}

		stream, err := directory.GetStream(streamID)
		if err != nil {
			logging.Error("failed to retrieve stream", "error", err)
			http.Error(w, "failed to retrieve stream", http.StatusInternalServerError)
			return
		}
		if stream == nil {
			http.Error(w, "stream not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+streamID+".ts\"")

		_, err = io.Copy(w, stream)
		if err != nil {
			logging.Error("failed to stream video", "error", err)
			http.Error(w, "failed to stream video", http.StatusInternalServerError)
			return
		}
	}
}

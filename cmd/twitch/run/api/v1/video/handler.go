package video

import (
	"net/http"
	"path/filepath"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/hls"
	"github.com/matthiasharzer/livebuffer/logging"
)

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamID := r.PathValue("streamID")
		cleanPath := filepath.Clean(r.URL.Path)
		filename := filepath.Base(cleanPath)

		filesDirectory, err := director.GetStreamFilesDirectory(streamID)
		if err != nil {
			logging.Error("failed to retrieve stream files directory", "error", err)
			http.Error(w, "failed to retrieve stream files", http.StatusInternalServerError)
			return
		}
		if filesDirectory == "" {
			http.Error(w, "stream not found", http.StatusNotFound)
			return
		}

		handler := hls.ServeHTTP(filesDirectory, filename)
		handler(w, r)
	}
}

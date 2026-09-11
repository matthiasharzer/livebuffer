package live

import (
	"net/http"
	"path/filepath"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/hls"
	"github.com/matthiasharzer/livebuffer/logging"
)

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cleanPath := filepath.Clean(r.URL.Path)
		filename := filepath.Base(cleanPath)

		filesDirectory, err := director.GetLiveStreamFilesDirectory()
		if err != nil {
			logging.Error("failed to retrieve live stream files directory", "error", err)
			http.Error(w, "failed to retrieve live stream files", http.StatusInternalServerError)
			return
		}
		if filesDirectory == "" {
			http.Error(w, "user is not live", http.StatusNotFound)
			return
		}

		handler := hls.ServeHTTP(filesDirectory, filename)
		handler(w, r)
	}
}

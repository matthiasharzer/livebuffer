package live

import (
	"net/http"
	"path/filepath"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/buffer/stream"
	"github.com/matthiasharzer/livebuffer/hls"
	"github.com/matthiasharzer/livebuffer/logging"
)

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cleanPath := filepath.Clean(r.URL.Path)
		filename := filepath.Base(cleanPath)

		streamInfo, err := director.GetLiveStreamInfo()
		if err != nil {
			logging.Error("failed to retrieve live stream", "error", err)
			http.Error(w, "failed to retrieve live stream", http.StatusInternalServerError)
			return
		}
		if streamInfo == nil {
			http.Error(w, "user is not live", http.StatusNotFound)
			return
		}
		filesDirectory := stream.FilesDirectory(streamInfo.Directory)

		handler := hls.ServeHTTP(filesDirectory, filename)
		handler(w, r)
	}
}

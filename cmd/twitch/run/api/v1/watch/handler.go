package watch

import (
	_ "embed"
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
)

//go:embed index.html
var indexHTML string

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !director.HasLiveStream() {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("No live stream available"))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(indexHTML))
	}
}

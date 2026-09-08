package live

import (
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
)

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp2t")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		clientChan, unsubscribe := director.LiveSubscribe()
		if clientChan == nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("No live stream available"))
			return
		}

		defer unsubscribe()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case chunk, ok := <-clientChan:
				if !ok {
					return
				}
				_, err := w.Write(chunk)
				if err != nil {
					return
				}
			}
		}
	}
}

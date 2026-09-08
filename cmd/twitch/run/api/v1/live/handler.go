package live

import (
	"net/http"

	"github.com/docker/go-units"
	"github.com/matthiasharzer/livebuffer/buffer"
)

const bufferSize = 1 * units.MiB

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp2t")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		liveManager, err := director.GetLiveManager()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Failed to get live manager"))
			return
		}
		if liveManager == nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("No live stream available"))
			return
		}

		clientChannel := make(chan []byte, bufferSize)
		liveManager.LiveSubscribe(clientChannel)

		defer liveManager.LiveUnsubscribe(clientChannel)

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case chunk, ok := <-clientChannel:
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

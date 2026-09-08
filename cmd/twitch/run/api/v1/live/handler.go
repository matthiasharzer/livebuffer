package live

import (
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
)

const channelBufferChunks = 250

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp2t")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, canFlush := w.(http.Flusher)
		if !canFlush {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

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

		clientChannel := make(chan []byte, channelBufferChunks)
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
				flusher.Flush()
			}
		}
	}
}

package live

import (
	"net/http"
	"time"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/logging"
)

const channelBufferChunks = 250
const chunkWriteDeadline = 15 * time.Second

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp2t")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, canFlush := w.(http.Flusher)
		if !canFlush {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}
		rc := http.NewResponseController(w)

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
		logging.Info("client connected to live stream")

		liveManager.LiveSubscribe(clientChannel)

		defer func() {
			liveManager.LiveUnsubscribe(clientChannel)
			logging.Info("client disconnected from live stream")
		}()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case chunk, ok := <-clientChannel:
				if !ok {
					return
				}

				err := rc.SetWriteDeadline(time.Now().Add(chunkWriteDeadline))
				if err != nil {
					logging.Error("failed to set deadline", "error", err)
					return
				}

				_, err = w.Write(chunk)
				if err != nil {
					logging.Warn("write error or timeout", "error", err)
					return
				}
				flusher.Flush()
			}
		}
	}
}

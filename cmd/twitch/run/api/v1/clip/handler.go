package clip

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
	"github.com/matthiasharzer/livebuffer/util/ioutil"
)

func Handler(directory *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamID := r.URL.Query().Get("stream_id")
		if streamID == "" {
			http.Error(w, "missing 'stream_id' query parameter", http.StatusBadRequest)
			return
		}
		startStr := r.URL.Query().Get("start")
		endStr := r.URL.Query().Get("end")

		if startStr == "" || endStr == "" {
			http.Error(w, "missing 'start' or 'end' query parameters (e.g., ?start=5m&end=15m)", http.StatusBadRequest)
			return
		}

		start, err := time.ParseDuration(startStr)
		if err != nil {
			http.Error(w, "invalid 'start' format. Use Go duration strings (e.g., 2m, 30s)", http.StatusBadRequest)
			return
		}
		end, err := time.ParseDuration(endStr)
		if err != nil {
			http.Error(w, "invalid 'end' format. Use Go duration strings (e.g., 2m, 30s)", http.StatusBadRequest)
			return
		}
		if start < 0 || end < 0 {
			http.Error(w, "'start' and 'end' must be non-negative durations", http.StatusBadRequest)
			return
		}
		if start >= end {
			http.Error(w, "'end' must be greater than 'start'", http.StatusBadRequest)
			return
		}

		clipInfo, clipReader, err := directory.GetClip(streamID, start, end)
		if err != nil {
			logging.Error("failed to retrieve stream", "error", err)
			http.Error(w, "failed to retrieve stream", http.StatusInternalServerError)
			return
		}
		if clipReader == nil {
			http.Error(w, "stream not found", http.StatusNotFound)
			return
		}
		defer funcutils.LogError(clipReader.Close, "failed to close stream")

		responseFileName := fmt.Sprintf("%s_%s_%s_to_%s.ts", clipInfo.Stream.BroadcasterUserName, streamID, startStr, endStr)
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+responseFileName+"\"")

		_, err = ioutil.CopyWithContext(r.Context(), w, clipReader)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				// expected when client disconnects while writing
				return
			}
			logging.Error("failed to stream video", "error", err)
			return
		}
	}
}

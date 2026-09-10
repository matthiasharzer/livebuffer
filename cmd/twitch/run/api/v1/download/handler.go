package download

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
)

func Handler(directory *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamID := r.URL.Query().Get("stream_id")
		if streamID == "" {
			http.Error(w, "missing 'stream_id' query parameter", http.StatusBadRequest)
			return
		}

		streamInfo, streamReader, err := directory.GetStream(streamID)
		if err != nil {
			logging.Error("failed to retrieve stream", "error", err)
			http.Error(w, "failed to retrieve stream", http.StatusInternalServerError)
			return
		}
		if streamReader == nil {
			http.Error(w, "stream not found", http.StatusNotFound)
			return
		}
		defer funcutils.LogError(streamReader.Close, "failed to close stream")

		fileName := fmt.Sprintf("%s_%s.ts", streamInfo.BroadcasterUserName, streamInfo.ID)

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", strconv.FormatInt(streamInfo.Size, 10))
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

		http.ServeContent(w, r, fileName, streamInfo.StartedAt, streamReader)
	}
}

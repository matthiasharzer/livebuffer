package list

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/buffer/vod/filter"
	"github.com/matthiasharzer/livebuffer/logging"
	"github.com/matthiasharzer/livebuffer/stream"
)

type StreamOrder string

const (
	StreamOrderChronological     StreamOrder = "chronological"
	StreamOrderAntichronological StreamOrder = "antichronological"
	StreamOrderAlphabetical      StreamOrder = "alphabetical"
)

type OrderFunc = func(a, b stream.Details) int

func getFilter(r *http.Request) filter.Func {
	username := r.URL.Query().Get("username")
	if username != "" {
		return filter.ByBroadcasterName(username)
	}
	return filter.All()
}

func getOrder(r *http.Request, w http.ResponseWriter) (StreamOrder, error) {
	strOrder := r.URL.Query().Get("order")
	switch StreamOrder(strOrder) {
	case "":
		return StreamOrderChronological, nil
	case StreamOrderChronological, StreamOrderAntichronological, StreamOrderAlphabetical:
		return StreamOrder(strOrder), nil
	default:
		http.Error(w, fmt.Sprintf("unknown order %s", strOrder), http.StatusBadRequest)
		return StreamOrderChronological, errors.New("received invalid stream order")
	}
}

func getOrderFunc(order StreamOrder) OrderFunc {
	switch order {
	case StreamOrderChronological:
		return func(a, b stream.Details) int {
			return a.StartedAt.Compare(b.StartedAt)
		}
	case StreamOrderAntichronological:
		return func(a, b stream.Details) int {
			return b.StartedAt.Compare(a.StartedAt)
		}
	case StreamOrderAlphabetical:
		return func(a, b stream.Details) int {
			return strings.Compare(a.Title, b.Title)
		}
	default:
		return func(a, b stream.Details) int {
			return 0
		}
	}
}

func Handler(director *buffer.Director) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		order, err := getOrder(r, w)
		if err != nil {
			logging.Warn("error while resolving order", "error", err)
			return
		}

		orderFunc := getOrderFunc(order)
		filterFunc := getFilter(r)
		streamsSeq := director.GetStreams(filterFunc)
		sortedStreams := slices.SortedFunc(streamsSeq, orderFunc)

		w.Header().Set("Content-Type", "application/json")
		responseStreams := make([]ResponseStream, 0, len(sortedStreams))
		for _, s := range sortedStreams {
			responseStreams = append(responseStreams, ResponseStream{
				ID:                   s.ID,
				Title:                s.Title,
				Size:                 humanize.Bytes(uint64(s.Size)),
				SizeBytes:            s.Size,
				Duration:             s.Duration.String(),
				DurationMilliseconds: s.Duration.Milliseconds(),
				StartedAt:            s.StartedAt,
				BroadcasterUserName:  s.BroadcasterUserName,
				StreamState:          string(s.StreamState),
			})
		}
		response := Response{
			Streams: responseStreams,
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logging.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

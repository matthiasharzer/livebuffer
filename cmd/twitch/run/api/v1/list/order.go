package list

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/matthiasharzer/livebuffer/stream"
)

type StreamOrder string

const (
	StreamOrderChronological       StreamOrder = "chronological"
	StreamOrderAntichronological   StreamOrder = "antichronological"
	StreamOrderAlphabetical        StreamOrder = "alphabetical"
	StreamOrderAlphabeticalReverse StreamOrder = "reverse-alphabetical"
	StreamOrderState               StreamOrder = "state"
)

type OrderFunc = func(a, b stream.Details) int

var orderFunctions = map[StreamOrder]OrderFunc{
	StreamOrderChronological: func(a, b stream.Details) int {
		return a.StartedAt.Compare(b.StartedAt)
	},
	StreamOrderAntichronological: func(a, b stream.Details) int {
		return b.StartedAt.Compare(a.StartedAt)
	},
	StreamOrderAlphabetical: func(a, b stream.Details) int {
		return strings.Compare(a.Title, b.Title)
	},
	StreamOrderAlphabeticalReverse: func(a, b stream.Details) int {
		return strings.Compare(b.Title, a.Title)
	},
	StreamOrderState: func(a, b stream.Details) int {
		return strings.Compare(string(a.StreamState), string(b.StreamState))
	},
}

func getOrderFunc(r *http.Request, w http.ResponseWriter) OrderFunc {
	strOrder := r.URL.Query().Get("order")
	if strOrder == "" {
		return orderFunctions[StreamOrderChronological]
	}
	orderFunc, ok := orderFunctions[StreamOrder(strOrder)]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown order %s", strOrder), http.StatusBadRequest)
		return nil
	}
	return orderFunc
}

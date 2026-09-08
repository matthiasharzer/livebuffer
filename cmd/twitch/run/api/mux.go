package api

import (
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/clip"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/download"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/list"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/live"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/watch"
	"github.com/matthiasharzer/livebuffer/twitch"
)

func GetMux(twitchAPI *twitch.Client, director *buffer.Director) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	mux.Handle("POST /api/v1/twitch-event-sub", twitchAPI.EventSubHTTPHandler())
	mux.HandleFunc("GET /api/v1/list", list.Handler(director))
	mux.HandleFunc("GET /api/v1/download", download.Handler(director))
	mux.HandleFunc("GET /api/v1/clip", clip.Handler(director))
	mux.HandleFunc("GET /api/v1/live", live.Handler(director))
	mux.HandleFunc("GET /api/v1/watch", watch.Handler(director))
	mux.HandleFunc("GET /", watch.Handler(director))

	return mux
}

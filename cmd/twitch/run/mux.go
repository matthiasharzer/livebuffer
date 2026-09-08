package run

import (
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/clip"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/download"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/list"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/live"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/ui"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/httputil"
)

func GetMux(twitchClient *twitch.Client, director *buffer.Director) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	mux.Handle("POST /api/v1/twitch-event-sub", twitchClient.EventSubHTTPHandler())
	mux.HandleFunc("GET /api/v1/list", list.Handler(director))
	mux.HandleFunc("GET /api/v1/download", download.Handler(director))
	mux.HandleFunc("GET /api/v1/clip", clip.Handler(director))
	mux.HandleFunc("GET /api/v1/live", live.Handler(director))

	mux.Handle("GET /api/", http.NotFoundHandler())
	mux.Handle("GET /",
		httputil.UseMiddleware(
			[]httputil.Middleware{httputil.GZIPMiddleware()},
			httputil.HandleStaticSite(ui.Content),
		),
	)

	return mux
}

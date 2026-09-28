package run

import (
	"fmt"
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/clip"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/download"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/list"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/live"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/video"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/ui"
	"github.com/matthiasharzer/livebuffer/util/httputil"
)

func GetMux(director *buffer.Director, usernames []string, eventSubHandler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	for _, username := range usernames {
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/list", username), list.Handler(director, username))
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/download", username), download.Handler(director))
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/clip", username), clip.Handler(director))
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/live/", username), live.Handler(director, username))
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/video/{streamID}/", username), video.Handler(director))
	}

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.Handle("POST /api/v1/twitch-event-sub", eventSubHandler)
	mux.Handle("GET /api/", http.NotFoundHandler())
	mux.Handle("GET /",
		httputil.UseMiddleware(
			[]httputil.Middleware{httputil.GZIPMiddleware()},
			httputil.HandleStaticSite(ui.Content),
		),
	)

	return mux
}

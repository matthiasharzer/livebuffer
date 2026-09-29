package run

import (
	"fmt"
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/clip"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/details"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/download"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/list"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/live"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/video"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/ui"
	"github.com/matthiasharzer/livebuffer/util/httputil"
)

const eventSubPath = "/api/v1/twitch-event-sub"

func GetMux(director *buffer.Director, usernames []string, eventSubHandler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/list", list.Handler(director))
	mux.HandleFunc("GET /api/v1/details/{streamID}", details.Handler(director))
	mux.HandleFunc("GET /api/v1/download/{streamID}", download.Handler(director))
	mux.HandleFunc("GET /api/v1/clip/{streamID}", clip.Handler(director))
	mux.HandleFunc("GET /api/v1/video/{streamID}/", video.Handler(director))

	for _, username := range usernames {
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/live/%s/", username), live.Handler(director, username))
	}

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.Handle(fmt.Sprintf("POST %s", eventSubPath), eventSubHandler)
	mux.Handle("GET /api/", http.NotFoundHandler())
	mux.Handle("GET /",
		httputil.UseMiddleware(
			[]httputil.Middleware{httputil.GZIPMiddleware()},
			httputil.HandleStaticSite(ui.Content),
		),
	)

	return mux
}

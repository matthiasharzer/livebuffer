package run

import (
	"fmt"
	"net/http"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/clip"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/download"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/list"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/live"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/api/v1/stream"
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run/ui"
	"github.com/matthiasharzer/livebuffer/twitch"
	"github.com/matthiasharzer/livebuffer/util/funcutils"
	"github.com/matthiasharzer/livebuffer/util/httputil"
)

type userContext struct {
	username     string
	twitchClient *twitch.Client
	director     *buffer.Director
}

func (u *userContext) Close() error {
	err := funcutils.CloseAll(u.twitchClient.Close, u.director.Close)
	return err
}

func GetMux(userContexts []userContext, eventSubHandler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	for _, context := range userContexts {
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/list", context.username), list.Handler(context.director))
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/download", context.username), download.Handler(context.director))
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/clip", context.username), clip.Handler(context.director))
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/live", context.username), live.Handler(context.director))
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/%s/{streamID}/video/", context.username), stream.Handler(context.director))
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

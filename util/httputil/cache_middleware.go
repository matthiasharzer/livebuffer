package httputil

import (
	"fmt"
	"net/http"
	"time"
)

func CacheMiddleware(maxAge time.Duration, hashFn func(r *http.Request) string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			hash := hashFn(r)
			if hash == "" {
				next(w, r)
				return
			}

			w.Header().Set("ETag", hash)
			w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(maxAge.Seconds())))

			match := r.Header.Get("If-None-Match")
			if match != "" && match == hash {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			next(w, r)
			return
		}
	}
}

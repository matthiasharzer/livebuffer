package httputil

import (
	"net/http"
	"strings"
)

func GZIPMiddleware() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next(w, r)
				return
			}

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Del("Content-Length")

			gzipResponseWriter := NewGZIPResponseWriter(w)
			defer func() {
				_ = gzipResponseWriter.Close()
			}()

			next(gzipResponseWriter, r)
		})
	}
}

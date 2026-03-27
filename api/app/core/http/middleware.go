package http

import (
	"log/slog"
	nethttp "net/http"
	"time"

	"localhost/app/core/log"
	"localhost/app/core/utils"
)

// MaxBytesReader returns middleware that limits the size of incoming
// request bodies to the given number of bytes. Requests that exceed
// the limit receive a 413 Request Entity Too Large response.
func MaxBytesReader(limit int64) func(nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			r.Body = nethttp.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	nethttp.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// RequestLogger returns middleware that assigns a request ID, logs each
// request with method, path, status, and duration, and sets the
// X-Request-ID response header.
func RequestLogger(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		requestID := utils.NewID()
		ctx := log.WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestID)

		rec := &statusRecorder{ResponseWriter: w, status: nethttp.StatusOK}
		start := time.Now()

		next.ServeHTTP(rec, r)

		slog.InfoContext(ctx, "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

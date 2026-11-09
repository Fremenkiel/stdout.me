package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/fremenkiel/stdout.me/internal/platform/metrics"
)

func LogRequestHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		metricsResponseWriter := metrics.NewResponseWriter(w)
		h.ServeHTTP(metricsResponseWriter, r)

		duration := time.Since(start)
		uri := r.URL.String()
		method := r.Method
		code := metricsResponseWriter.Code
		writtenLen := metricsResponseWriter.WrittenLen

		logHttpRequest(method, uri, duration, code, writtenLen)
	})
}

func logHttpRequest(method, uri string, duration time.Duration, code, writtenLen int) {
	LogInfo(
		"logger/middleware",
		"incoming request",
		slog.String("method", method),
		slog.String("uri", uri),
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.Int("code", code),
		slog.Int("written_len", writtenLen),
		)
}

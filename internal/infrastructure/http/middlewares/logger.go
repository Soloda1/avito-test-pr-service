package middlewares

import (
	"avito-test-pr-service/internal/infrastructure/logger"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLoggerMiddleware(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			requestID := middleware.GetReqID(r.Context())
			entry := log.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)

			if r.URL.RawQuery != "" {
				entry = entry.With(slog.String("rawQuery", r.URL.RawQuery))
			}

			entry.Info("request started")

			t1 := time.Now()

			next.ServeHTTP(ww, r)

			entry.Info("request completed",
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.String("duration", time.Since(t1).String()),
			)
		}

		return http.HandlerFunc(fn)
	}
}

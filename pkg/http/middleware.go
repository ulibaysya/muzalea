package http

import (
	"log/slog"
	"net/http"
)

func MaxBytesMiddleware(n int64) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.MaxBytesHandler(h, n)
	}
}

func AccessMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return access(logger, next)
	}
}

func access(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger.LogAttrs(ctx, slog.LevelInfo, "serving",
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
		)

		statusCodeCapturer := statusCodeCapturer{w, 200}
		next.ServeHTTP(&statusCodeCapturer, r)

		logger.LogAttrs(ctx, slog.LevelInfo, "done serving",
			slog.Int("statusCode", statusCodeCapturer.statusCode),
		)
	})

}

type statusCodeCapturer struct {
	http.ResponseWriter
	statusCode int
}

func (c *statusCodeCapturer) WriteHeader(statusCode int) {
	c.statusCode = statusCode
	c.ResponseWriter.WriteHeader(statusCode)
}

func PanicRecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return panicRecovery(logger, next)
	}
}

func panicRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recoveryMsg := recover()
			if recoveryMsg != nil {
				logger.LogAttrs(r.Context(), slog.LevelError, "caught panic",
					slog.Any("recoveryMessage", recoveryMsg),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				)
				ErrorJSON500(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

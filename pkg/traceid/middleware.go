package traceid

import (
	"log/slog"
	"net/http"

	pkgrand "github.com/ulibaysya/muzalea/pkg/rand"
)

func HTTPMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := pkgrand.Base64URL(32)
			logger.Info("serving trace ID", slog.String("traceID", id))
			next.ServeHTTP(w, r.WithContext(ContextWith(r.Context(), id)))
		})
	}
}

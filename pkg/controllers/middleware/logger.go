package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := slog.String("id", middleware.GetReqID(r.Context()))
			ctx := context.WithValue(r.Context(), LoggerRequestKey, logger.With(id))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

var (
	LoggerRequestKey = &contextKey{"requestLogger"}
	GetRequestLogger = GetFromContext[slog.Logger](LoggerRequestKey)
)

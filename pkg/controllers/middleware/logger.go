package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

var LoggerRequestKey = &contextKey{"loggerRequest"}

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := slog.String("id", middleware.GetReqID(r.Context()))
			ctx := context.WithValue(r.Context(), LoggerRequestKey, logger.With(id))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetRequestLogger(ctx context.Context) (*slog.Logger, error) {
	obj := ctx.Value(LoggerRequestKey)
	if obj == nil {
		return nil, errors.New("Expected *slog.Logger")
	}
	req, ok := obj.(*slog.Logger)
	if !ok {
		return nil, errors.New("Object stored in request context couldn't convert to *slog.Logger")
	}
	return req, nil
}

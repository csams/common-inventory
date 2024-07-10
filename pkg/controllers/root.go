package controllers

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	slogchi "github.com/samber/slog-chi"

	"gorm.io/gorm"

	authnapi "github.com/csams/common-inventory/pkg/authn/api"
	cimw "github.com/csams/common-inventory/pkg/controllers/middleware"
	eventingapi "github.com/csams/common-inventory/pkg/eventing/api"
)

func NewRootHandler(db *gorm.DB, authenticator authnapi.Authenticator, eventingManager eventingapi.Manager, log *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(slogchi.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	r.Get("/healthz", Ready)

	r.With(
		cimw.Logger(log),
		cimw.Authentication(authenticator),
		render.SetContentType(render.ContentTypeJSON),
	).
		Route("/api/inventory/v1.0", func(r chi.Router) {

			// These type specific controllers can be simplified with go generics, but we can't settle on a
			// set of standard event types or common handling logic across all resource types.
			r.Mount("/resources", NewResourceController(db, eventingManager, log).Routes())
		})

	return r
}

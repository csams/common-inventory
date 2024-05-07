package controllers

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	slogchi "github.com/samber/slog-chi"

	"gorm.io/gorm"
)

func NewRootHandler(db *gorm.DB, log *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(slogchi.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/healthz", Ready)

	r.Route("/api/inventory/v1.0", func(r chi.Router) {
		r.Mount("/resources", NewResourceController(db, log).Routes())
		r.Mount("/linux-hosts", NewHostController(db, log).Routes())
		r.Mount("/k8s-clusters", NewClusterController(db, log).Routes())
	})

	return r
}

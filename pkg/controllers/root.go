package controllers

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/csams/common-inventory/pkg/models"

	slogchi "github.com/samber/slog-chi"

	"gorm.io/gorm"
)

func NewHandler(db *gorm.DB, log *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(slogchi.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/healthz", Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/resources", NewCRUD[models.Resource](db, log, WithoutMutatingVerbs).Routes())
		r.Mount("/hosts", NewCRUD(db, log, AsResourceType[models.Host]("host", "hosts")).Routes())
		r.Mount("/clusters", NewCRUD(db, log, AsResourceType[models.Cluster]("cluster", "clusters")).Routes())
	})

	return r
}

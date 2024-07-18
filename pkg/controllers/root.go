package controllers

import (
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	slogchi "github.com/samber/slog-chi"

	"gorm.io/gorm"

	authnapi "github.com/csams/common-inventory/pkg/authn/api"
	authzapi "github.com/csams/common-inventory/pkg/authz/api"
	mw "github.com/csams/common-inventory/pkg/controllers/middleware"
	eventingapi "github.com/csams/common-inventory/pkg/eventing/api"
)

func NewRootHandler(db *gorm.DB, authenticator authnapi.Authenticator, authorizer authzapi.Authorizer, eventingManager eventingapi.Manager, log *slog.Logger) chi.Router {
	basePath := "/api/inventory/v1alpha1"

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(slogchi.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	r.Get("/healthz", Ready)

	r.With(
		mw.Logger(log),
		mw.Authentication(authenticator),
		render.SetContentType(render.ContentTypeJSON),
	).
		Route(basePath, func(r chi.Router) {
			r.Mount("/resources/hosts", NewResourceController(fmt.Sprintf("%s/resources/hosts", basePath), "host", db, authorizer, eventingManager, log).Routes())
			r.Mount("/resources/clusters", NewResourceController(fmt.Sprintf("%s/resources/clusters", basePath), "cluster", db, authorizer, eventingManager, log).Routes())
			r.Mount("/resources/acm-policies", NewResourceController(fmt.Sprintf("%s/resources/acm-policies", basePath), "acm-policy", db, authorizer, eventingManager, log).Routes())
		})

	return r
}

package controllers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/csams/common-inventory/pkg/models"
)

type Controller[T any, P interface {
	*T
	models.TypedResource
}] struct {
	Db  *gorm.DB
	Log *slog.Logger

	resourceType         string
	resourceTypePlural   string
	includeMutatingVerbs bool
}

func NewCRUD[M any, P interface {
	*M
	models.TypedResource
}](db *gorm.DB, log *slog.Logger, opts ...Option[M, P]) *Controller[M, P] {
	c := &Controller[M, P]{
		Db:                   db,
		Log:                  log,
		includeMutatingVerbs: true,
	}

	for _, opt := range opts {
		c = opt(c)
	}

	return c
}

func (c Controller[T, P]) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(Pagination).Get("/", c.List)

	if c.includeMutatingVerbs {
		r.Post("/", c.Create)
	}

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", c.Get)
		if c.includeMutatingVerbs {
			r.Put("/", c.Update)
			r.Delete("/", c.Delete)
		}
	})

	return r
}

func (c *Controller[T, P]) List(w http.ResponseWriter, r *http.Request) {
	pagination, err := PaginationRequestFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var model T
	var count int64
	if err := c.Db.Model(&model).Count(&count).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db := c.Db.Scopes(pagination.Filter)

	var results []T
	if err := db.Preload(clause.Associations).Find(&results).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &PagedResponse[T]{
		PagedReponseMetadata: PagedReponseMetadata{
			Page:  pagination.Page,
			Size:  len(results),
			Total: count,
		},
		Items: results,
	}

	render.JSON(w, r, resp)
}

func (c *Controller[T, P]) Create(w http.ResponseWriter, r *http.Request) {
	var model T
	if err := render.Decode(r, &model); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var p P = &model
	p.SetResourceType(c.resourceType)

	if err := c.Db.Create(p).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p.SetHref(fmt.Sprintf("/api/v1/%s/%d", c.resourceTypePlural, p.GetId()))

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Model(p).Updates(p).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, p)
}

func (c *Controller[T, P]) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model T
	if err := c.Db.Preload(clause.Associations).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	render.JSON(w, r, &model)
}

func (c *Controller[T, P]) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model T
	if err := c.Db.Preload(clause.Associations).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := render.Decode(r, &model); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&model).Save(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller[M, P]) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model M
	if err := c.Db.Delete(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type Option[M any, P interface {
	*M
	models.TypedResource
}] func(*Controller[M, P]) *Controller[M, P]

func WithoutMutatingVerbs[M any, P interface {
	*M
	models.TypedResource
}](c *Controller[M, P]) *Controller[M, P] {
	c.includeMutatingVerbs = false
	return c
}

func AsResourceType[M any, P interface {
	*M
	models.TypedResource
}](t string, p string) Option[M, P] {
	return func(c *Controller[M, P]) *Controller[M, P] {
		c.resourceType = t
		c.resourceTypePlural = p
		return c
	}
}

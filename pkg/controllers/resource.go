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

	"github.com/csams/common-inventory/pkg/controllers/middleware"
	"github.com/csams/common-inventory/pkg/models"
)

type ResourceController struct {
	Db  *gorm.DB
	Log *slog.Logger
}

func NewResourceController(db *gorm.DB, log *slog.Logger) *ResourceController {
	return &ResourceController{
		Db:  db,
		Log: log,
	}
}

func (c ResourceController) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(middleware.Pagination).Get("/", c.List)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", c.Get)
	})

	return r
}

func (c *ResourceController) List(w http.ResponseWriter, r *http.Request) {
	pagination, err := middleware.PaginationRequestFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var model models.Resource
	var count int64
	if err := c.Db.Model(&model).Count(&count).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db := c.Db.Scopes(pagination.Filter)

	var results []models.Resource
	if err := db.Preload(clause.Associations).Find(&results).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var output []models.ResourceOut
	for _, result := range results {
		out := models.ResourceOut{
			Resource: result,
			Href:     fmt.Sprintf("/api/inventory/v1.0/resources/%d", result.ID),
		}
		output = append(output, out)
	}

	resp := &middleware.PagedResponse[models.ResourceOut]{
		PagedReponseMetadata: middleware.PagedReponseMetadata{
			Page:  pagination.Page,
			Size:  len(results),
			Total: count,
		},
		Items: output,
	}

	render.JSON(w, r, resp)
}

func (c *ResourceController) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Resource
	if err := c.Db.Preload(clause.Associations).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

    out := models.ResourceOut{
        Resource: model,
        Href:     fmt.Sprintf("/api/inventory/v1.0/resources/%d", model.ID),
    }
	render.JSON(w, r, out)
}

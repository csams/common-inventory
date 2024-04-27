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

	"github.com/csams/common-inventory/pkg/controllers/middleware"
	"github.com/csams/common-inventory/pkg/models"
)

type HostController struct {
	Db  *gorm.DB
	Log *slog.Logger
}

func NewHostController(db *gorm.DB, log *slog.Logger) *HostController {
	return &HostController{
		Db:  db,
		Log: log,
	}
}

func (c HostController) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(middleware.Pagination).Get("/", c.List)
	r.With(middleware.Reporter).Post("/", c.Create)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", c.Get)
		r.With(middleware.Reporter).Put("/", c.Update)
		r.With(middleware.Reporter).Delete("/", c.Delete)
	})

	return r
}

func (c *HostController) List(w http.ResponseWriter, r *http.Request) {
	pagination, err := middleware.PaginationRequestFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var model models.Host
	var count int64
	if err := c.Db.Model(&model).Count(&count).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db := c.Db.Scopes(pagination.Filter)

	var results []models.Host
	if err := db.Preload("Metadata.Reporters").Preload("Metadata.Tags").Find(&results).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var output []models.HostOut
	for _, r := range results {
		out := models.HostOut{
			Metadata:   models.ResourceOut{Resource: r.Metadata},
			HostCommon: r.HostCommon,
		}
		out.Metadata.Href = fmt.Sprintf("/api/inventory/v1.0/hosts/%d", r.ID)
		output = append(output, out)
	}

	resp := &middleware.PagedResponse[models.HostOut]{
		PagedReponseMetadata: middleware.PagedReponseMetadata{
			Page:  pagination.Page,
			Size:  len(results),
			Total: count,
		},
		Items: output,
	}

	render.JSON(w, r, resp)
}

func (c *HostController) Create(w http.ResponseWriter, r *http.Request) {
	reporter, err := middleware.ReporterFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input models.HostIn
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata := models.Resource{
		DisplayName:  input.Metadata.DisplayName,
		Tags:         input.Metadata.Tags,
		ResourceType: "linux-host",
		Reporters: []models.Reporter{
			{
				Name: reporter.Name,
				Type: reporter.Type,
				URL:  reporter.URL,

				Created: input.Metadata.ReporterTime,
				Updated: input.Metadata.ReporterTime,
			},
		},
	}

	model := models.Host{
		Metadata:   metadata,
		HostCommon: input.HostCommon,
	}

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := &models.HostOut{
		Metadata:   models.ResourceOut{Resource: model.Metadata},
		HostCommon: model.HostCommon,
	}
	out.Metadata.Href = fmt.Sprintf("/api/inventory/v1.0/hosts/%d", model.ID)

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, out)
}

func (c *HostController) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Host
	if err := c.Db.Preload("Metadata.Reporters").Preload("Metadata.Tags").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	out := models.HostOut{
		Metadata:   models.ResourceOut{Resource: model.Metadata},
		HostCommon: model.HostCommon,
	}

	out.Metadata.Href = fmt.Sprintf("api/inventory/v1.0/hosts/%d", model.ID)
	render.JSON(w, r, &model)
}

func (c *HostController) Update(w http.ResponseWriter, r *http.Request) {
	reporter, err := middleware.ReporterFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input models.HostIn
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Host
	if err := c.Db.Preload("Metadata.Reporters").Preload("Metadata.Tags").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	model.Metadata.DisplayName = input.Metadata.DisplayName
	model.Metadata.Tags = input.Metadata.Tags
	model.Metadata.UpdatedAt = input.Metadata.ReporterTime

	model.HostCommon = input.HostCommon
	for _, r := range model.Metadata.Reporters {
		if r.Name == reporter.Name {
			r.Updated = input.Metadata.ReporterTime
		}
	}

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&model).Save(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *HostController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Host
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

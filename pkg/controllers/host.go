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
	eventingapi "github.com/csams/common-inventory/pkg/eventing/api"
	"github.com/csams/common-inventory/pkg/models"
)

type HostController struct {
	Db              *gorm.DB
	EventingManager eventingapi.Manager
	Log             *slog.Logger
}

func NewHostController(db *gorm.DB, eventingManager eventingapi.Manager, log *slog.Logger) *HostController {
	return &HostController{
		Db:  db,
        EventingManager: eventingManager,
		Log: log,
	}
}

func (c HostController) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(middleware.Pagination).Get("/", c.List)
	r.Post("/", c.Create)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", c.Get)
		r.Put("/", c.Update)
		r.Delete("/", c.Delete)
	})

	return r
}

func (c *HostController) List(w http.ResponseWriter, r *http.Request) {
	pagination, err := middleware.GetPaginationRequest(r.Context())
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
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input models.HostIn
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model := models.Host{
		Metadata: models.Resource{
			DisplayName:  input.Metadata.DisplayName,
			Tags:         input.Metadata.Tags,
			ResourceType: "linux-host",
			Reporters: []models.Reporter{
				{
					Name: identity.Principal,
					Type: identity.Type,
					URL:  identity.Href,

					Created: input.Metadata.ReporterTime,
					Updated: input.Metadata.ReporterTime,
				},
			},
		},
		HostCommon: input.HostCommon,
	}

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    producer, _ := c.EventingManager.Lookup(identity, "linux-host", model.ID)
    evt := &eventingapi.Event[*models.Host]{
        EventType: "Create",
        ResourceType: "linux-host",
        Object: &model,
    }
    producer.Produce(evt)

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
	identity, err := middleware.GetIdentity(r.Context())
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
		if r.Name == identity.Principal {
			r.Updated = input.Metadata.ReporterTime
		}
	}

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&model).Save(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    producer, _ := c.EventingManager.Lookup(identity, "linux-host", models.IDType(id))
    evt := &eventingapi.Event[*models.Host]{
        EventType: "Update",
        ResourceType: "linux-host",
        Object: &model,
    }
    producer.Produce(evt)

	w.WriteHeader(http.StatusNoContent)
}

func (c *HostController) Delete(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

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

    producer, _ := c.EventingManager.Lookup(identity, "linux-host", models.IDType(id))
    evt := &eventingapi.Event[*models.Host]{
        EventType: "Delete",
        ResourceType: "linux-host",
        Object: &model,
    }
    producer.Produce(evt)

	w.WriteHeader(http.StatusNoContent)
}

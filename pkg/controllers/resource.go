package controllers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/csams/common-inventory/pkg/controllers/middleware"
	cerrors "github.com/csams/common-inventory/pkg/errors"
	eventingapi "github.com/csams/common-inventory/pkg/eventing/api"
	"github.com/csams/common-inventory/pkg/models"
)

type ResourceController struct {
	Db              *gorm.DB
	EventingManager eventingapi.Manager
	Log             *slog.Logger
}

func NewResourceController(db *gorm.DB, em eventingapi.Manager, log *slog.Logger) *ResourceController {
	return &ResourceController{
		Db:              db,
		EventingManager: em,
		Log:             log,
	}
}

func (c ResourceController) Routes() chi.Router {
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

func (c *ResourceController) List(w http.ResponseWriter, r *http.Request) {
	_, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	pagination, err := middleware.GetPaginationRequest(r.Context())
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
	_, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

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

func (c *ResourceController) Create(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input models.ResourceIn
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

    if errs := input.Validate(); errs != nil {
        http.Error(w, cerrors.NewAggregate(errs).Error(), http.StatusBadRequest)
        return
    }

	model := models.Resource{
		DisplayName:  input.DisplayName,
		Tags:         input.Tags,
		ResourceType: input.ResourceType,
		Reporters: []models.Reporter{
			{
				Name: identity.Principal,
				Type: identity.Type,
				URL:  identity.Href,

				Created: input.ReporterTime,
				Updated: input.ReporterTime,
                ResourceIdAlias: input.ResourceIdAlias,
			},
		},
		Data: datatypes.JSON(input.Data),
	}

	if err := c.Db.Create(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: handle eventing errors
	// TODO: Update the Object that's sent.  This is going to be what we actually emit.
	producer, _ := c.EventingManager.Lookup(identity, input.ResourceType, model.ID)
	evt := &eventingapi.Event[models.Resource]{
		EventType:    "Create",
		ResourceType: input.ResourceType,
		Object:       model,
	}
	producer.Produce(evt)

	out := &models.ResourceOut{
		Resource: model,
		Href:     fmt.Sprintf("/api/inventory/v1.0/resources/%d", model.ID),
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, out)
}

func (c *ResourceController) Update(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input models.ResourceIn
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Resource
	if err := c.Db.Preload("Metadata.Reporters").Preload("Metadata.Tags").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	model.DisplayName = input.DisplayName
	model.Tags = input.Tags
	model.UpdatedAt = input.ReporterTime

	for _, r := range model.Reporters {
		if r.Name == identity.Principal {
			r.Updated = input.ReporterTime
		}
	}

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&model).Save(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: handle eventing errors
	// TODO: Update the Object that's sent.  This is going to be what we actually emit.
	producer, _ := c.EventingManager.Lookup(identity, input.ResourceType, model.ID)
	evt := &eventingapi.Event[models.Resource]{
		EventType:    "Update",
		ResourceType: input.ResourceType,
		Object:       model,
	}
	producer.Produce(evt)

	w.WriteHeader(http.StatusNoContent)
}

func (c *ResourceController) Delete(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Resource
	if err := c.Db.Delete(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// TODO: handle eventing errors
	// TODO: Update the Object that's sent.  This is going to be what we actually emit.
	producer, _ := c.EventingManager.Lookup(identity, model.ResourceType, model.ID)
	evt := &eventingapi.Event[models.Resource]{
		EventType:    "Update",
		ResourceType: model.ResourceType,
		Object:       model,
	}
	producer.Produce(evt)

	w.WriteHeader(http.StatusNoContent)
}

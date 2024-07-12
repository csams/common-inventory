package controllers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/csams/common-inventory/pkg/authn/api"
	"github.com/csams/common-inventory/pkg/controllers/middleware"
	cerrors "github.com/csams/common-inventory/pkg/errors"
	eventingapi "github.com/csams/common-inventory/pkg/eventing/api"
	"github.com/csams/common-inventory/pkg/models"
)

type ResourceController struct {
	BasePath        string
	Db              *gorm.DB
	EventingManager eventingapi.Manager
	Log             *slog.Logger
}

func NewResourceController(
	basePath string,
	db *gorm.DB,
	em eventingapi.Manager,
	log *slog.Logger) *ResourceController {
	return &ResourceController{
		BasePath:        basePath,
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

	var output []*models.ResourceOut
	for _, result := range results {
		r := result
		href := fmt.Sprintf("%s/%d", c.BasePath, result.ID)
		out := models.NewResourceOutput(&r, href)
		output = append(output, out)
	}

	resp := &middleware.PagedResponse[*models.ResourceOut]{
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

	href := fmt.Sprintf("%s/%d", c.BasePath, model.ID)
	out := models.NewResourceOutput(&model, href)
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

	model := c.CreateResourceFromInput(&input, identity)

	if err := c.Db.Create(model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if c.EventingManager != nil {
		// TODO: handle eventing errors
		// TODO: Update the Object that's sent.  This is going to be what we actually emit.
		producer, _ := c.EventingManager.Lookup(identity, model.ResourceType, model.ID)
		evt := &eventingapi.Event[*models.Resource]{
			EventType:    "Create",
			ResourceType: model.ResourceType,
			Object:       model,
		}
		producer.Produce(evt)
	}

	href := fmt.Sprintf("%s/%d", c.BasePath, model.ID)
	out := models.NewResourceOutput(model, href)

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

	db := c.Db.Preload("Reporters").Preload("Tags")

	var model models.Resource
	if err := db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	c.UpdateResourceFromInput(&input, &model, identity)

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if c.EventingManager != nil {
		// TODO: handle eventing errors
		// TODO: Update the Object that's sent.  This is going to be what we actually emit.
		producer, _ := c.EventingManager.Lookup(identity, model.ResourceType, model.ID)
		evt := &eventingapi.Event[*models.Resource]{
			EventType:    "Update",
			ResourceType: model.ResourceType,
			Object:       &model,
		}
		producer.Produce(evt)
	}

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

	if c.EventingManager != nil {
		// TODO: handle eventing errors
		// TODO: Update the Object that's sent.  This is going to be what we actually emit.
		producer, _ := c.EventingManager.Lookup(identity, model.ResourceType, model.ID)
		evt := &eventingapi.Event[*models.Resource]{
			EventType:    "Update",
			ResourceType: model.ResourceType,
			Object:       &model,
		}
		producer.Produce(evt)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ResourceController) CreateResourceFromInput(input *models.ResourceIn, identity *api.Identity) *models.Resource {
	return &models.Resource{
		// CreatedAt and UpdatedAt will be updated automatically by gorm
		DisplayName:  input.DisplayName,
		Tags:         input.Tags,
		ResourceType: strings.ToLower(input.ResourceType),
		Data:         datatypes.JSON(input.Data),
		Reporters: []models.Reporter{
			{
				Created: input.LocalTime,
				Updated: input.LocalTime,

				Name: identity.Principal,
				Type: identity.Type,
				URL:  identity.Href,

				ConsoleHref: input.ConsoleHref,
				ApiHref:     input.ApiHref,

				LocalId: input.LocalId,
			},
		},
	}
}

func (c *ResourceController) UpdateResourceFromInput(input *models.ResourceIn, model *models.Resource, identity *api.Identity) {
	model.DisplayName = input.DisplayName
	model.Tags = input.Tags
	model.UpdatedAt = input.LocalTime
	model.Data = datatypes.JSON(input.Data)

	found := false
	for i := range model.Reporters {
		r := &model.Reporters[i]
		if r.Name == identity.Principal {
			found = true

			r.Updated = input.LocalTime

			r.Type = identity.Type
			r.URL = identity.Href

			r.ConsoleHref = input.ConsoleHref
			r.ApiHref = input.ApiHref

			r.LocalId = input.LocalId
		}
	}

	if !found {
		reporter := models.Reporter{
			Created: input.LocalTime,
			Updated: input.LocalTime,

			Name: identity.Principal,
			Type: identity.Type,
			URL:  identity.Href,

			ConsoleHref: input.ConsoleHref,
			ApiHref:     input.ApiHref,

			LocalId: input.LocalId,
		}
		model.Reporters = append(model.Reporters, reporter)
	}

}

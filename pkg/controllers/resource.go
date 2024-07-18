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

	authnapi "github.com/csams/common-inventory/pkg/authn/api"
	authzapi "github.com/csams/common-inventory/pkg/authz/api"
	"github.com/csams/common-inventory/pkg/controllers/middleware"
	cerrors "github.com/csams/common-inventory/pkg/errors"
	eventingapi "github.com/csams/common-inventory/pkg/eventing/api"
	"github.com/csams/common-inventory/pkg/models"
)

type ResourceController struct {
	BasePath        string
	ResourceType    string
	Db              *gorm.DB
	Authorizer      authzapi.Authorizer
	EventingManager eventingapi.Manager
	Log             *slog.Logger
}

func NewResourceController(
	basePath string,
	resourceType string,
	db *gorm.DB,
	authorizer authzapi.Authorizer,
	em eventingapi.Manager,
	log *slog.Logger) *ResourceController {
	return &ResourceController{
		BasePath:        basePath,
		ResourceType:    resourceType,
		Db:              db,
		Authorizer:      authorizer,
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
	if err := db.Preload(clause.Associations).Where("resource_type = ?", c.ResourceType).Find(&results).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var output []*models.ResourceOut
	for _, result := range results {
		r := result
		href := fmt.Sprintf("%s/%d", c.BasePath, result.ID)
		out := models.NewResourceOut(&r, href)
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

	rawId := chi.URLParam(r, "id")
	var model models.Resource
	id, err := strconv.ParseInt(rawId, 10, 64)
	if err == nil {
		if err := c.Db.Preload(clause.Associations).First(&model, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		// if the id isn't an integer, it's the special format
		parts := strings.Split(rawId, ":")
		if len(parts) != 4 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if parts[0] != "hcrn" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		reporterType, reporterInstanceId, localResourceId := parts[1], parts[2], parts[3]
		if err := c.Db.
			Preload(clause.Associations).
			Joins("join reporter_data on reporter_data.resource_id = resources.id").
			Where("reporter_data.reporter_id = ? and reporter_data.reporter_type = ? and reporter_data.local_resource_id = ? and resources.resource_type = ?", reporterInstanceId, reporterType, localResourceId, c.ResourceType).
			First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	href := fmt.Sprintf("%s/%d", c.BasePath, model.ID)
	out := models.NewResourceOut(&model, href)
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

	model, err := c.CreateResourceFromInput(&input, identity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.Db.Create(model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if c.EventingManager != nil {
		// TODO: handle eventing errors
		// TODO: Update the Object that's sent.  This is going to be what we actually emit.
		producer, _ := c.EventingManager.Lookup(identity, model)
		evt := &eventingapi.Event{
			EventType:    "Create",
			ResourceType: c.ResourceType,
			Object:       model,
		}
		producer.Produce(r.Context(), evt)
	}

	href := fmt.Sprintf("%s/%d", c.BasePath, model.ID)
	out := models.NewResourceOut(model, href)

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
		return
	}

	db := c.Db.Preload("ReporterData")

	var model models.Resource
	if err := db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = c.UpdateResourceFromInput(&input, &model, identity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if c.EventingManager != nil {
		// TODO: handle eventing errors
		// TODO: Update the Object that's sent.  This is going to be what we actually emit.
		producer, _ := c.EventingManager.Lookup(identity, &model)
		evt := &eventingapi.Event{
			EventType:    "Update",
			ResourceType: c.ResourceType,
			Object:       &model,
		}
		producer.Produce(r.Context(), evt)
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
		producer, _ := c.EventingManager.Lookup(identity, &model)
		evt := &eventingapi.Event{
			EventType:    "Update",
			ResourceType: c.ResourceType,
			Object:       &model,
		}
		producer.Produce(r.Context(), evt)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ResourceController) CreateResourceFromInput(input *models.ResourceIn, identity *authnapi.Identity) (*models.Resource, error) {
	var reporterType string
	if len(identity.Type) > 0 {
		reporterType = identity.Type
	} else if len(input.ReporterType) > 0 {
		reporterType = input.ReporterType
	} else {
		return nil, fmt.Errorf("ReporterType must not be empty.")
	}

	var count int64
	c.Db.Model(&models.ReporterData{}).
		Where("reporter_id = ? and reporter_type = ? and local_resource_id = ?", identity.Principal, reporterType, input.LocalResourceId).
		Count(&count)

	if count > 0 {
		return nil, fmt.Errorf("Resource for instance %s of ReporterType %s already exists", identity.Principal, reporterType)
	}

	return &models.Resource{
		// CreatedAt and UpdatedAt will be updated automatically by gorm
		DisplayName:  input.DisplayName,
		ResourceType: strings.ToLower(c.ResourceType),
		Workspace:    input.Workspace,
		ReporterData: []models.ReporterData{{
			ReporterID: identity.Principal,

			Created: input.LocalTime,
			Updated: input.LocalTime,

			LocalResourceId: input.LocalResourceId,
			ReporterType:    reporterType,
			ReporterVersion: input.ReporterVersion,

			ConsoleHref: input.ConsoleHref,
			ApiHref:     input.ApiHref,

			Data: datatypes.JSON(input.Data),
		}},
	}, nil
}

func (c *ResourceController) UpdateResourceFromInput(input *models.ResourceIn, model *models.Resource, identity *authnapi.Identity) error {
	model.DisplayName = input.DisplayName
	if input.Workspace != nil {
		model.Workspace = input.Workspace
	}

	found := false
	for i := range model.ReporterData {
		r := &model.ReporterData[i]
		if r.ReporterID == identity.Principal {
			found = true

			r.Updated = input.LocalTime
			r.ReporterVersion = input.ReporterVersion
			r.Data = datatypes.JSON(input.Data)

			r.ConsoleHref = input.ConsoleHref
			r.ApiHref = input.ApiHref

			r.LocalResourceId = input.LocalResourceId
		}
	}

	if !found {
		reporter := models.ReporterData{
			ReporterID: identity.Principal,

			Created: input.LocalTime,
			Updated: input.LocalTime,

			LocalResourceId: input.LocalResourceId,
			ReporterType:    identity.Type,
			ReporterVersion: input.ReporterVersion,

			ConsoleHref: input.ConsoleHref,
			ApiHref:     input.ApiHref,

			Data: datatypes.JSON(input.Data),
		}
		model.ReporterData = []models.ReporterData{reporter}
	}
	return nil
}

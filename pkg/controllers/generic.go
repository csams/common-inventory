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
	cerrors "github.com/csams/common-inventory/pkg/errors"
	eventingapi "github.com/csams/common-inventory/pkg/eventing/api"
	"github.com/csams/common-inventory/pkg/models"
)

type Controller[I models.Input, M models.Model, O models.Output] struct {
	BasePath        string
	Transformer     models.Transformer[I, M, O]
	Db              *gorm.DB
	Preloads        []string
	EventingManager eventingapi.Manager
	Log             *slog.Logger
}

func NewController[I models.Input, M models.Model, O models.Output](
	basePath string,
	processor models.Transformer[I, M, O],
	preloads []string,
	db *gorm.DB,
	em eventingapi.Manager,
	log *slog.Logger) *Controller[I, M, O] {
	return &Controller[I, M, O]{
		BasePath:        basePath,
		Db:              db,
		Transformer:     processor,
		Preloads:        preloads,
		EventingManager: em,
		Log:             log,
	}
}

func (c Controller[I, M, O]) Routes() chi.Router {
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

func (c *Controller[I, M, O]) List(w http.ResponseWriter, r *http.Request) {
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

	model := c.Transformer.NewModel()
	var count int64
	if err := c.Db.Model(&model).Count(&count).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db := c.Db.Scopes(pagination.Filter)

	var results []M
	if err := db.Preload(clause.Associations).Find(&results).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var output []O
	for _, result := range results {
		out := c.Transformer.ToOutput(result)
		out.SetHref(fmt.Sprintf("%s/%d", c.BasePath, result.GetId()))
		output = append(output, out)
	}

	resp := &middleware.PagedResponse[O]{
		PagedReponseMetadata: middleware.PagedReponseMetadata{
			Page:  pagination.Page,
			Size:  len(results),
			Total: count,
		},
		Items: output,
	}

	render.JSON(w, r, resp)
}

func (c *Controller[I, M, O]) Get(w http.ResponseWriter, r *http.Request) {
	_, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	model := c.Transformer.NewModel()
	if err := c.Db.Preload(clause.Associations).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	out := c.Transformer.ToOutput(model)
	out.SetHref(fmt.Sprintf("%s/%d", c.BasePath, model.GetId()))
	render.JSON(w, r, out)
}

func (c *Controller[I, M, O]) Create(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	input := c.Transformer.NewInput()
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if errs := input.Validate(); errs != nil {
		http.Error(w, cerrors.NewAggregate(errs).Error(), http.StatusBadRequest)
		return
	}

	model := c.Transformer.Create(input, identity)

	if err := c.Db.Create(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if c.EventingManager != nil {
		// TODO: handle eventing errors
		// TODO: Update the Object that's sent.  This is going to be what we actually emit.
		producer, _ := c.EventingManager.Lookup(identity, model.GetResourceType(), model.GetId())
		evt := &eventingapi.Event[M]{
			EventType:    "Create",
			ResourceType: model.GetResourceType(),
			Object:       model,
		}
		producer.Produce(evt)
	}

	out := c.Transformer.ToOutput(model)
	out.SetHref(fmt.Sprintf("%s/%d", c.BasePath, model.GetId()))

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, out)
}

func (c *Controller[I, M, O]) Update(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	input := c.Transformer.NewInput()
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	db := c.Db
	for _, preload := range c.Preloads {
		db = db.Preload(preload)
	}

	model := c.Transformer.NewModel()
	if err := db.First(model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	c.Transformer.Update(input, model, identity)

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Model(model).Save(model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if c.EventingManager != nil {
		// TODO: handle eventing errors
		// TODO: Update the Object that's sent.  This is going to be what we actually emit.
		producer, _ := c.EventingManager.Lookup(identity, model.GetResourceType(), model.GetId())
		evt := &eventingapi.Event[M]{
			EventType:    "Update",
			ResourceType: model.GetResourceType(),
			Object:       model,
		}
		producer.Produce(evt)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller[I, M, O]) Delete(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	model := c.Transformer.NewModel()
	if err := c.Db.Delete(model, id).Error; err != nil {
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
		producer, _ := c.EventingManager.Lookup(identity, model.GetResourceType(), model.GetId())
		evt := &eventingapi.Event[M]{
			EventType:    "Update",
			ResourceType: model.GetResourceType(),
			Object:       model,
		}
		producer.Produce(evt)
	}

	w.WriteHeader(http.StatusNoContent)
}

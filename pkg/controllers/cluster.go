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

type ClusterController struct {
	Db  *gorm.DB
	Log *slog.Logger
}

func NewClusterController(db *gorm.DB, log *slog.Logger) *ClusterController {
	return &ClusterController{
		Db:  db,
		Log: log,
	}
}

func (c ClusterController) Routes() chi.Router {
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

func (c *ClusterController) List(w http.ResponseWriter, r *http.Request) {
	pagination, err := middleware.GetPaginationRequest(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var model models.Cluster
	var count int64
	if err := c.Db.Model(&model).Count(&count).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db := c.Db.Scopes(pagination.Filter)

	var results []models.Cluster
	if err := db.Preload("Metadata.Reporters").Preload("Metadata.Tags").Find(&results).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var output []models.ClusterOut
	for _, r := range results {
		out := models.ClusterOut{
			Metadata:      models.ResourceOut{Resource: r.Metadata},
			ClusterCommon: r.ClusterCommon,
		}
		out.Metadata.Href = fmt.Sprintf("/api/inventory/v1.0/clusters/%d", r.ID)
		output = append(output, out)
	}

	resp := &middleware.PagedResponse[models.ClusterOut]{
		PagedReponseMetadata: middleware.PagedReponseMetadata{
			Page:  pagination.Page,
			Size:  len(results),
			Total: count,
		},
		Items: output,
	}

	render.JSON(w, r, resp)
}

func (c *ClusterController) Create(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input models.ClusterIn
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model := &models.Cluster{
		Metadata: models.Resource{
			DisplayName:  input.Metadata.DisplayName,
			Tags:         input.Metadata.Tags,
			ResourceType: "k8s-cluster",
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
		ClusterCommon: input.ClusterCommon,
	}

	if err := c.Db.Create(model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := &models.ClusterOut{
		Metadata:      models.ResourceOut{Resource: model.Metadata},
		ClusterCommon: model.ClusterCommon,
	}
	out.Metadata.Href = fmt.Sprintf("/api/inventory/v1.0/clusters/%d", model.ID)

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, out)
}

func (c *ClusterController) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Cluster
	if err := c.Db.Preload("Metadata.Reporters").Preload("Metadata.Tags").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	out := models.ClusterOut{
		Metadata:      models.ResourceOut{Resource: model.Metadata},
		ClusterCommon: model.ClusterCommon,
	}

	out.Metadata.Href = fmt.Sprintf("api/inventory/v1.0/cluster/%d", model.ID)
	render.JSON(w, r, &model)
}

func (c *ClusterController) Update(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input models.ClusterIn
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Cluster
	if err := c.Db.Preload("Metadata.Reporters").Preload("Metadata.Reporters").First(&model, id).Error; err != nil {
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

	model.ClusterCommon = input.ClusterCommon
	for _, r := range model.Metadata.Reporters {
		if r.Name == identity.Principal {
			r.Updated = input.Metadata.ReporterTime
		}
	}

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&model).Save(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ClusterController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Cluster
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

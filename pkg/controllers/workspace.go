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

	authnapi "github.com/csams/common-inventory/pkg/authn/api"
	authzapi "github.com/csams/common-inventory/pkg/authz/api"
	"github.com/csams/common-inventory/pkg/controllers/middleware"
	cerrors "github.com/csams/common-inventory/pkg/errors"
	"github.com/csams/common-inventory/pkg/models"
)

type WorkspaceController struct {
	BasePath   string
	Db         *gorm.DB
	Authorizer authzapi.Authorizer
	Log        *slog.Logger
}

func NewWorkspaceController(
	basePath string,
	db *gorm.DB,
	authorizer authzapi.Authorizer,
	log *slog.Logger) *WorkspaceController {
	return &WorkspaceController{
		BasePath:   basePath,
		Db:         db,
		Authorizer: authorizer,
		Log:        log,
	}
}

func (c WorkspaceController) Routes() chi.Router {
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

func (c *WorkspaceController) List(w http.ResponseWriter, r *http.Request) {
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

	var model models.Workspace
	var count int64
	if err := c.Db.Model(&model).Count(&count).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db := c.Db.Scopes(pagination.Filter)

	var results []models.Workspace
	if err := db.Preload(clause.Associations).Find(&results).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var output []*models.WorkspaceOut
	for _, result := range results {
		href := fmt.Sprintf("%s/%d", c.BasePath, result.ID)
		out := models.NewWorkspaceOut(&result, href)
		output = append(output, out)
	}

	resp := &middleware.PagedResponse[*models.WorkspaceOut]{
		PagedReponseMetadata: middleware.PagedReponseMetadata{
			Page:  pagination.Page,
			Size:  len(results),
			Total: count,
		},
		Items: output,
	}

	render.JSON(w, r, resp)
}

func (c *WorkspaceController) Get(w http.ResponseWriter, r *http.Request) {
	_, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Workspace
	if err := c.Db.Preload(clause.Associations).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	href := fmt.Sprintf("%s/%d", c.BasePath, model.ID)
	out := models.NewWorkspaceOut(&model, href)
	render.JSON(w, r, out)
}

func (c *WorkspaceController) Create(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input models.WorkspaceIn
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if errs := input.Validate(); errs != nil {
		http.Error(w, cerrors.NewAggregate(errs).Error(), http.StatusBadRequest)
		return
	}

	model := c.CreateWorkspaceFromInput(&input, identity)

	if err := c.Db.Create(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	href := fmt.Sprintf("%s/%d", c.BasePath, model.ID)
	out := models.NewWorkspaceOut(model, href)

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, out)
}

func (c *WorkspaceController) Update(w http.ResponseWriter, r *http.Request) {
	identity, err := middleware.GetIdentity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input models.WorkspaceIn
	if err := render.Decode(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Workspace
	if err := c.Db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	c.UpdateWorkspaceFromInput(&input, &model, identity)

	if err := c.Db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&model).Save(&model).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *WorkspaceController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var model models.Workspace
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

func (c *WorkspaceController) CreateWorkspaceFromInput(input *models.WorkspaceIn, identity *authnapi.Identity) *models.Workspace {
	return &models.Workspace{
		DisplayName: input.DisplayName,
	}
}

func (c *WorkspaceController) UpdateWorkspaceFromInput(input *models.WorkspaceIn, model *models.Workspace, identity *authnapi.Identity) {
	model.DisplayName = input.DisplayName
}

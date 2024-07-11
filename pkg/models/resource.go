package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/csams/common-inventory/pkg/authn/api"
	"gorm.io/datatypes"
)

type IDType int64

// ResourceIn is a REST API mixin for specific resource types
type ResourceIn struct {
	DisplayName string
	// Allow reporters to specify when something was created or updated.  Don't allow them to have control
	// over announcing their type or Id.  Reporter type and Id should be derived or otherwise communicated
	// from a signed token.
	LocalTime time.Time

	// LocalId is the identifier given to the resource by the reporter
	LocalId string

	ConsoleHref string
	ApiHref     string

	ResourceType string
	Data         json.RawMessage

	Tags []ResourceTag
}

func (r *ResourceIn) Validate() []error {
	var errs []error

	if len(r.DisplayName) == 0 {
		errs = append(errs, errors.New("DisplayName must not be empty"))
	}

	if len(r.LocalId) == 0 {
		errs = append(errs, errors.New("LocalId must not be empty"))
	}

	if len(r.ResourceType) == 0 {
		errs = append(errs, errors.New("ResourceType must not be empty"))
	}

	if len(r.Data) == 0 {
		errs = append(errs, errors.New("Data must not be empty"))
	}

	return errs
}

// ResourceOut is a REST API mixin for specific resource types
type ResourceOut struct {
	*Resource
	Href string
}

func (r *ResourceOut) SetHref(href string) {
	r.Href = href
}

type Resource struct {
	ID        IDType `gorm:"primaryKey" json:"-"` // don't send this in the REST API
	CreatedAt time.Time
	UpdatedAt time.Time `json:"LastUpdatedAt"`

	DisplayName  string `gorm:"not null"`
	ResourceType string `gorm:"not null"`
	Workspace    string

	Data datatypes.JSON `json:"Data" gorm:"not null"`

	Reporters []Reporter `gorm:"many2many:resource_reporters"`
	Tags      []ResourceTag
}

func (r *Resource) GetId() IDType {
	return r.ID
}

func (r *Resource) GetResourceType() string {
	return r.ResourceType
}

// Reporter is a lower inventory communicating a creation or update of a resource
type Reporter struct {
	Name string `gorm:"primaryKey"`

	// "CreatedAt" and "UpdatedAt" have special meaning to gorm.  We want these to represent when the lower
	// inventory created or updated the resource, not when the fields are updated in the common inventory. The
	// lower inventory can update these fields indirectly through ResourceIn.ReporterTime.
	Created time.Time `json:"CreatedAt"`
	Updated time.Time `json:"LastUpdatedAt"`

	Type string `gorm:"not null"`
	URL  string `gorm:"not null"`

	ConsoleHref string
	ApiHref     string

	// This is the primary key assigned to the resource *by the reporter*.
	LocalId string `gorm:"not null"`
}

// ResourceTag is a way for a resource to be tagged and queried for cross cutting concerns
type ResourceTag struct {
	ID IDType `gorm:"primaryKey" json:"-"` // don't send this in the REST API

	// A Resource may have many ResourceTags
	ResourceID IDType `json:"-"`

	Namespace string `gorm:"not null"`
	Key       string `gorm:"not null"`
	Value     string
}

type ResourceProcessor struct{}

func NewResourceTransformer() *ResourceProcessor {
	return &ResourceProcessor{}
}

func (p *ResourceProcessor) NewInput() *ResourceIn {
	return &ResourceIn{}
}

func (p *ResourceProcessor) NewModel() *Resource {
	return &Resource{}
}

func (p *ResourceProcessor) Create(input *ResourceIn, identity *api.Identity) *Resource {
	return &Resource{
		// CreatedAt and UpdatedAt will be updated automatically by gorm
		DisplayName:  input.DisplayName,
		Tags:         input.Tags,
		ResourceType: input.ResourceType,
		Data:         datatypes.JSON(input.Data),
		Reporters: []Reporter{
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

func (p *ResourceProcessor) Update(input *ResourceIn, model *Resource, identity *api.Identity) {
	model.DisplayName = input.DisplayName
	model.Tags = input.Tags
	model.UpdatedAt = input.LocalTime
	model.Data = datatypes.JSON(input.Data)

	found := false
	for _, r := range model.Reporters {
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
		reporter := Reporter{
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

func (p *ResourceProcessor) ToOutput(r *Resource) *ResourceOut {
	return &ResourceOut{
		Resource: r,
	}
}

var _ Transformer[*ResourceIn, *Resource, *ResourceOut] = &ResourceProcessor{}

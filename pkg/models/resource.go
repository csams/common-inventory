package models

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/datatypes"
)

type IDType int64

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
	Spec         json.RawMessage
	Status       json.RawMessage

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

	if len(r.Spec) == 0 {
		errs = append(errs, errors.New("Spec must not be empty"))
	}

	return errs
}

type Resource struct {
	ID        IDType `gorm:"primaryKey" json:"-"` // don't send this in the REST API
	CreatedAt time.Time
	UpdatedAt time.Time `json:"LastUpdatedAt"`

	DisplayName  string `gorm:"not null"`
	ResourceType string `gorm:"not null"`
	WorkspaceID  *IDType

	Spec   datatypes.JSON `json:"Spec" gorm:"not null"`
	Status datatypes.JSON `json:"Status"`

	Reporters []Reporter
	Tags      []ResourceTag
}

// Reporter is a lower inventory communicating a creation or update of a resource
type Reporter struct {
	ID IDType `json:"-" gorm:"primaryKey"`

	Name string

	// "CreatedAt" and "UpdatedAt" have special meaning to gorm.  We want these to represent when the lower
	// inventory created or updated the resource, not when the fields are updated in the common inventory. The
	// lower inventory can update these fields indirectly through ResourceIn.ReporterTime.
	Created time.Time `json:"CreatedAt"`
	Updated time.Time `json:"LastUpdatedAt"`

	Type string `gorm:"not null"`
	URL  string `gorm:"not null"`

	ConsoleHref string
	ApiHref     string

	ResourceID IDType `json:"-"`

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

// ResourceOut is a REST API mixin for specific resource types
type ResourceOut struct {
	*Resource
	Href string
}

func NewResourceOutput(r *Resource, href string) *ResourceOut {
	return &ResourceOut{
		Resource: r,
		Href:     href,
	}
}

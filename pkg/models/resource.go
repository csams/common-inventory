package models

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/datatypes"
)

type IDType int64

type ResourceIn struct {
	// LocalResourceId is the identifier assigned to the resource within the reporter's system.
	LocalResourceId string

	// Is resource this a k8s cluster?  A host?
	ResourceType string

	// Human readable name
	DisplayName string

	// Allow reporters to specify when the resource was created or updated.  Don't allow them to have control
	// over announcing their type or Id.  Reporter type and Reporter Id should be inferred (perhaps from the
	// identity data of the caller) or otherwise communicated to common inventory.
	LocalTime time.Time

	// Workspace is the ID of the workspace to which the resource is associated.
	Workspace *string

	// URLs to where to access the resource
	ConsoleHref string
	ApiHref     string

	// This is the type of the Data blob below.  It specifies whether this is an OCM cluster, an ACM cluster,
	// etc.  It seems reasonable to infer the value from the caller's identity data, but it's not clear that's
	// *always* the case.  So, allow it to be passed explicitly and then log a warning or something if the value
	// doesn't match the inferred type.
	ReporterType string

	// The version of the reporter.
	ReporterVersion string

	// This should be data about the resource type that is specific to the ReporterType.
	Data json.RawMessage
}

func (r *ResourceIn) Validate() []error {
	var errs []error

	if len(r.DisplayName) == 0 {
		errs = append(errs, errors.New("Resource DisplayName must not be empty"))
	}

	if len(r.ResourceType) == 0 {
		errs = append(errs, errors.New("Resource ResourceType must not be empty"))
	}

	if len(r.LocalResourceId) == 0 {
		errs = append(errs, errors.New("ReporterData LocalResourceId must not be empty"))
	}

	if len(r.ReporterType) == 0 {
		errs = append(errs, errors.New("Resource ReporterType must not be empty"))
	}

	if len(r.Data) == 0 {
		errs = append(errs, errors.New("ReporterData Data must not be empty"))
	}
	if r.Data == nil || len(r.Data) == 0 {
		errs = append(errs, errors.New("Resource Data must not be empty"))
	}

	return errs
}

type ReporterData struct {
	// ReporterID should be populated from the Identity of the caller.  e.g. if this is an ACM reporter, *which* ACM
	// instance is it?
	ReporterID string `gorm:"primaryKey"`

	// This is necessary to satisfy gorm so the collection in the Resource model works.
	ResourceID IDType `json:"-"`

	// This is the type of the Data blob below.  It specifies whether this is an OCM cluster, an ACM cluster,
	// etc.  It seems reasonable to infer the value from the caller's identity data, but it's not clear that's
	// *always* the case.  So, allow it to be passed explicitly and then log a warning or something if the value
	// doesn't match the inferred type.
	ReporterType string `gorm:"primaryKey"`

	Created time.Time
	Updated time.Time

	// LocalResourceId is the identifier assigned to the resource within the reporter's system.
	LocalResourceId string `gorm:"primaryKey"`

	// The version of the reporter.
	ReporterVersion string

	// pointers to where to access the resource
	ConsoleHref string
	ApiHref     string

	// the data of the resource.  Validation cannot be compiled in.
	Data datatypes.JSON
}

type Resource struct {
	ID        IDType `gorm:"primaryKey" json:"-"` // don't send this in the REST API
	CreatedAt time.Time
	UpdatedAt time.Time `json:"LastUpdatedAt"`

	DisplayName  string `gorm:"not null"`
	ResourceType string `gorm:"not null"`
	Workspace    *string

	// ReporterData is a map from ReporterType to the reporter's representation of the resource.
	ReporterData []ReporterData
}

type ResourceOut struct {
	*Resource
	Href string
}

func NewResourceOut(r *Resource, href string) *ResourceOut {
	return &ResourceOut{
		Resource: r,
		Href:     href,
	}
}

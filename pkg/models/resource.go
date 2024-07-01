package models

import "time"

type IDType int64

// ResourceIn is a REST API mixin for specific resource types
type ResourceIn struct {
	DisplayName string
	// Allow reporters to specify when something was created or updated.  Don't allow them to have control
	// over announcing their type or Id.  Reporter type and Id should be derived or otherwise communicated
	// from a signed token.
	ReporterTime    time.Time
	ConsoleHref     string
	ApiHref         string
	ResourceIdAlias string

	Tags []ResourceTag
}

// ResourceOut is a REST API mixin for specific resource types
type ResourceOut struct {
	Resource
	Href string
}

// Resource is common to entities that have a resource type in the database model
type Resource struct {
	ID        IDType `gorm:"primaryKey" json:"-"` // don't send this in the REST API
	CreatedAt time.Time
	UpdatedAt time.Time `json:"LastUpdatedAt"`

	DisplayName  string `gorm:"not null"`
	ResourceType string `gorm:"not null"`
	Workspace    string

	Reporters []Reporter
	Tags      []ResourceTag
}

// Reporter is a lower inventory communicating a creation or update of a resource
type Reporter struct {
	Name string `gorm:"primaryKey"`

	// "CreatedAt" and "UpdatedAt" have special meaning to gorm.  We want these to represent when the lower
	// inventory created or updated the resource, not when the fields are updated in the common inventory. The
	// lower inventory can update these fields indirectly through ResourceIn.ReporterTime.
	Created time.Time `json:"CreatedAt"`
	Updated time.Time `json:"LastUpdatedAt"`

	// A Resource may have many Reporters.
	// This tells GORM to set up a "has many" relationship from the Resource side.
	ResourceID IDType `json:"-"` // don't expose the ResourceID in the REST API

	Type string `gorm:"not null"`
	URL  string `gorm:"not null"`

	ConsoleHref string
	ApiHref     string

	// This is the primary key assigned to the resource *by the reporter*.
	ResourceIdAlias string `gorm:"not null"`
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

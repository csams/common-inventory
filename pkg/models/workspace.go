package models

import (
	"errors"
	"time"
)

type WorkspaceIn struct {
	DisplayName       string
	ParentWorkspaceId *IDType
}

func (r *WorkspaceIn) Validate() []error {
	var errs []error

	if len(r.DisplayName) == 0 {
		errs = append(errs, errors.New("DisplayName must not be empty"))
	}

	return errs
}

type WorkspaceOut struct {
	*Workspace
	Href string
}

type Workspace struct {
	ID        IDType `gorm:"primaryKey" json:"-"` // don't send this in the REST API
	CreatedAt time.Time
	UpdatedAt time.Time `json:"LastUpdatedAt"`

	DisplayName string `gorm:"not null"`

	ParentWorkspaceId *IDType
}

func NewWorkspaceOut(r *Workspace, href string) *WorkspaceOut {
	return &WorkspaceOut{
		Workspace: r,
		Href:      href,
	}
}

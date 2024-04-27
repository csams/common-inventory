package models

import (
	"time"
)

// ClusterIn is part of REST API
type ClusterIn struct {
	Metadata ResourceIn
	ClusterCommon
}

// ClusterOut is part of REST API
type ClusterOut struct {
	Metadata ResourceOut
	ClusterCommon
}

// Cluster is part of the database model
type Cluster struct {
	ClusterCommon

	ID        IDType    `gorm:"primaryKey" json:"-"` // don't send this in the REST API
	CreatedAt time.Time `json:"-"`                   // don't send this in the REST API
	UpdatedAt time.Time `json:"-"`                   // don't send this in the REST API

	ResourceID IDType   `json:"-"` // don't send this in the REST API
	Metadata   Resource `gorm:"foreignKey:ResourceID;constraint:OnDelete:CASCADE"`
}

// ClusterCommon is common to the REST API and the database model
type ClusterCommon struct {
	ExternalClusterId string
	CloudProviderId   string
	ApiServer         string
}

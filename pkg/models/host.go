package models

import (
	"time"
)

// HostIn is part of the REST API
type HostIn struct {
	Metadata ResourceIn
	HostCommon
}

// HostOut is part of the REST API
type HostOut struct {
	Metadata ResourceOut
	HostCommon
}

// Host is the database model
type Host struct {
	HostCommon

	ID        IDType    `gorm:"primaryKey" json:"-"` // don't send this in the REST API
	CreatedAt time.Time `json:"-"`                   // don't send this in the REST API
	UpdatedAt time.Time `json:"-"`                   // don't send this in the REST API

	ResourceID IDType   `json:"-"` // don't send this in the REST API
	Metadata   Resource `gorm:"foreignKey:ResourceID;constraint:OnDelete:CASCADE"`
}

// HostCommon is common to the REST API and the database model
type HostCommon struct {
	BiosUuid              string
	Fqdn                  string
	InsightsId            string
	ProviderId            string
	ProviderType          string
	SatelliteId           string
	SubscriptionManagerId string
}

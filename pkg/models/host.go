package models

type Host struct {
	Common

	ResourceID int64    `json:"-"`
	Metadata   Resource `gorm:"foreignKey:ResourceID;constraint:OnDelete:CASCADE"`

	BiosUuid              string
	Fqdn                  string
	InsightsId            string
	ProviderId            string
	ProviderType          string
	SatelliteId           string
	SubscriptionManagerId string
}

func (h *Host) SetResourceType(s string) {
	h.Metadata.ResourceType = s
}

func (h *Host) GetResourceType() string {
	return h.Metadata.ResourceType
}

func (h *Host) SetHref(s string) {
	h.Metadata.Href = s
}

func (h *Host) GetHref() string {
	return h.Metadata.Href
}

func (h *Host) GetId() int64 {
	return h.ID
}

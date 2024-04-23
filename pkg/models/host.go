package models

type Host struct {
	Common

	ResourceID int64    `json:"-"`
	Metadata   Resource `gorm:"foreignKey:ResourceID;constraint:OnDelete:CASCADE" json:"metadata"`

	BiosUuid              string `gorm:"bios_uuid" json:"bios_uuid"`
	FQDN                  string `gorm:"fqdn" json:"fqdn"`
	InsightsId            string `gorm:"insights_id" json:"insights_id"`
	ProviderId            string `gorm:"provider_id" json:"provider_id"`
	ProviderType          string `gorm:"provider_type" json:"provider_type"`
	SatelliteId           string `gorm:"satellite_id" json:"satellite_id"`
	SubscriptionManagerId string `gorm:"subcription_manager_id" json:"subcription_manager_id"`
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

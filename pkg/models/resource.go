package models

type Resource struct {
	Common

	Href         string `gorm:"href" json:"href"`
	DisplayName  string `gorm:"display_name" json:"display_name"`
	Reporter     string `gorm:"reporter" json:"reporter"`
	ResourceType string `gorm:"resource_type" json:"resource_type"`
}

func (r *Resource) SetResourceType(s string) {
	r.ResourceType = s
}

func (r *Resource) GetResourceType() string {
	return r.ResourceType
}

func (r *Resource) SetHref(s string) {
	r.Href = s
}

func (r *Resource) GetHref() string {
	return r.Href
}

func (r *Resource) GetId() int64 {
	return r.ID
}

type TypedResource interface {
	SetResourceType(string)
	GetResourceType() string
	SetHref(string)
	GetHref() string
	GetId() int64
}

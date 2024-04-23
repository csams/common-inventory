package models

type Cluster struct {
	Common

	ResourceID int64    `json:"-"`
	Metadata   Resource `gorm:"foreignKey:ResourceID;constraint:OnDelete:CASCADE"`
	ApiServer  string
}

func (c *Cluster) SetResourceType(s string) {
	c.Metadata.ResourceType = s
}

func (c *Cluster) GetResourceType() string {
	return c.Metadata.ResourceType
}

func (c *Cluster) SetHref(s string) {
	c.Metadata.Href = s
}

func (c *Cluster) GetHref() string {
	return c.Metadata.Href
}

func (c *Cluster) GetId() int64 {
	return c.ID
}

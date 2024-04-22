package models

import (
	"time"

	"gorm.io/gorm"
)

type Host struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Hostname  string         `gorm:"hostname" json:"hostname"`
}

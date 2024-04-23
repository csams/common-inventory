package models

import (
	"time"
)

type Common struct {
	ID        int64     `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

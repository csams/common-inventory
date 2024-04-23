package models

import (
	"time"
)

type Common struct {
	ID        int64 `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

package storage

import (
	"gorm.io/gorm"

	"github.com/csams/common-inventory/pkg/models"
)

// Migrate the schemas
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.Host{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&models.Cluster{}); err != nil {
		return err
	}
	return nil
}

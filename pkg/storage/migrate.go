package storage

import (
	"gorm.io/gorm"

	"github.com/csams/common-inventory/pkg/models"
)

// Migrate the schemas
// See https://gorm.io/docs/migration.html
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.Resource{}, &models.Reporter{}, &models.ResourceTag{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&models.Cluster{}, &models.Host{}); err != nil {
		return err
	}
	return nil
}

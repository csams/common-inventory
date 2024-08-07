package models

import (
	"gorm.io/gorm"
)

// Migrate the tables
// See https://gorm.io/docs/migration.html
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Resource{}, &ReporterData{}); err != nil {
		return err
	}
	return nil
}

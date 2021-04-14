package model

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	types := []interface{}{
		&ShortURL{},
		&ShortURLMetrics{},
	}
	if err := db.AutoMigrate(types...); err != nil {
		return err
	}
	return nil
}
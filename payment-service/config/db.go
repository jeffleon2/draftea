package config

import (
	"fmt"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"
)

func (db *DB) GormConnect() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		db.HOST, db.USER, db.PASSWORD, db.NAME, db.PORT, db.SSLMODE,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

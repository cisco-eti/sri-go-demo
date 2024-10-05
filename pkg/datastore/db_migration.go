package datastore

import (
	"gorm.io/gorm"

	"github.com/cisco-eti/sre-go-helloworld/pkg/models"
)

type migrateFunc func(db *gorm.DB) error

func Migrate(db *gorm.DB) error {
	migrations := []migrateFunc{
		migratePet,
		migrateUser,
		migrateSession,
	}

	for _, m := range migrations {
		err := m(db)
		if err != nil {
			return err
		}
	}

	return nil
}

func migratePet(db *gorm.DB) error {
	return db.AutoMigrate(&models.Pet{})
}

func migrateUser(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{})
}

func migrateSession(db *gorm.DB) error {
	return db.AutoMigrate(&models.Session{})
}

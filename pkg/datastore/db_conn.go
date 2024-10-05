package datastore

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"

	"github.com/cisco-eti/sre-go-helloworld/pkg/config"
	"github.com/cisco-eti/sre-go-helloworld/pkg/utils"
)

func OpenDB() (*gorm.DB, error) {
	dsn, err := config.ReadDBconfig()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          utils.DatabaseName,
		RefreshInterval: 15,
		StartServer:     false,
	}))
	if err != nil {
		return nil, err
	}

	return db, nil
}

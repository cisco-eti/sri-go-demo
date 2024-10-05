package pet

import (
	"gorm.io/gorm"

	etilogger "wwwin-github.cisco.com/eti/sre-go-logger"
)

type Pet struct {
	log *etilogger.Logger
	db  *gorm.DB
}

func New(l *etilogger.Logger, db *gorm.DB) *Pet {
	return &Pet{log: l, db: db}
}

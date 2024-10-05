package device

import (
	"gorm.io/gorm"

	etilogger "wwwin-github.cisco.com/eti/sre-go-logger"
)

type Device struct {
	log *etilogger.Logger
	db  *gorm.DB
}

func New(l *etilogger.Logger, db *gorm.DB) *Device {
	return &Device{log: l, db: db}
}

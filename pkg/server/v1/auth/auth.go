package auth

import (
	"gorm.io/gorm"

	"github.com/cisco-eti/sre-go-helloworld/pkg/idpadapter"
	etilogger "wwwin-github.cisco.com/eti/sre-go-logger"
)

type Auth struct {
	log        *etilogger.Logger
	db         *gorm.DB
	idpAdapter *idpadapter.IdentityProviderAdapter
}

func New(l *etilogger.Logger, db *gorm.DB,
	ipa *idpadapter.IdentityProviderAdapter) *Auth {
	return &Auth{
		log:        l,
		db:         db,
		idpAdapter: ipa,
	}
}

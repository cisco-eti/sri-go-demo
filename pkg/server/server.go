package server

import (
	"gorm.io/gorm"

	"github.com/cisco-eti/sre-go-helloworld/pkg/idpadapter"
	v1auth "github.com/cisco-eti/sre-go-helloworld/pkg/server/v1/auth"
	v1device "github.com/cisco-eti/sre-go-helloworld/pkg/server/v1/device"
	v1pet "github.com/cisco-eti/sre-go-helloworld/pkg/server/v1/pet"
	etilogger "wwwin-github.cisco.com/eti/sre-go-logger"
)

type Server struct {
	log      *etilogger.Logger
	v1auth   *v1auth.Auth
	v1device *v1device.Device
	v1pet    *v1pet.Pet
}

func New(l *etilogger.Logger, db *gorm.DB,
	ipa *idpadapter.IdentityProviderAdapter) *Server {
	return &Server{
		log:      l,
		v1auth:   v1auth.New(l, db, ipa),
		v1device: v1device.New(l, db),
		v1pet:    v1pet.New(l, db),
	}
}

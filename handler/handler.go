package handler

import (
	log "github.com/inconshreveable/log15"
	"github.com/music-gang/music-gang-api/app/service"
)

type ServiceHandler struct {
	ContractSearchService service.ContractSearchService
	AuthSearchService     service.AuthSearchService
	UserSearchService     service.UserSearchService
	VmCallableService     service.VmCallableService
	JWTService            service.JWTService

	Logger log.Logger
}

// NewServiceHandler creates a new ServiceHandler.
func NewServiceHandler() *ServiceHandler {
	return NewServiceHandlerWithLogger(log.Root())
}

func NewServiceHandlerWithLogger(logger log.Logger) *ServiceHandler {
	return &ServiceHandler{
		Logger: logger,
	}
}

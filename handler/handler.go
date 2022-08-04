package handler

import "github.com/music-gang/music-gang-api/app/service"

type ServiceHandler struct {
	ContractSearchService service.ContractSearchService
	AuthSearchService     service.AuthSearchService
	UserSearchService     service.UserSearchService
	VmCallableService     service.VmCallableService
	JWTService            service.JWTService

	LogService service.LogService
}

// NewServiceHandler creates a new ServiceHandler.
func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{}
}

package mock

import (
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.VmCallableService = (*VmCallableService)(nil)

type VmCallableService struct {
	*UserService
	*AuthService
	*FuelTankService
	*ContractService
	*ExecutorService
}

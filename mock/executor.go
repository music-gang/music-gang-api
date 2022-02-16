package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/service"
)

type ExecutorService struct {
	ExecContractFn func(ctx context.Context, contractRef *service.ContractCall) (res interface{}, err error)
}

func (e *ExecutorService) ExecContract(ctx context.Context, contractRef *service.ContractCall) (res interface{}, err error) {
	if e.ExecContractFn == nil {
		panic("ExecContractFn is not defined")
	}
	return e.ExecContractFn(ctx, contractRef)
}

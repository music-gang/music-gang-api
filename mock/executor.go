package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.ContractExecutorService = (*ExecutorService)(nil)

type ExecutorService struct {
	ExecContractFn func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error)
}

func (e *ExecutorService) ExecContract(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
	if e.ExecContractFn == nil {
		panic("ExecContractFn is not defined")
	}
	return e.ExecContractFn(ctx, opt)
}

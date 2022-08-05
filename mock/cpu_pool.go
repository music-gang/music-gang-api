package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.CPUsPoolService = (*CPUsPoolService)(nil)

type CPUsPoolService struct {
	AcquireCoreFn func(ctx context.Context, call service.VmCallable) (release func(), err error)
}

func (s *CPUsPoolService) AcquireCore(ctx context.Context, call service.VmCallable) (release func(), err error) {
	if s.AcquireCoreFn == nil {
		panic("AcquireCoreFn is not defined")
	}
	return s.AcquireCoreFn(ctx, call)
}

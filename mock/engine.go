package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.EngineService = (*EngineService)(nil)

type EngineService struct {
	ExecContractFn func(ctx context.Context, revision *entity.Revision) (res interface{}, err error)
	IsRunningFn    func() bool
	PauseFn        func() error
	ResumeFn       func() error
	StateFn        func() entity.VmState
	StopFn         func() error
}

func (e *EngineService) ExecContract(ctx context.Context, revision *entity.Revision) (res interface{}, err error) {
	if e.ExecContractFn == nil {
		panic("ExecContractFn is not defined")
	}
	return e.ExecContractFn(ctx, revision)
}

func (e *EngineService) IsRunning() bool {
	if e.IsRunningFn == nil {
		panic("IsRunningFn is not defined")
	}
	return e.IsRunningFn()
}

func (e *EngineService) Pause() error {
	if e.PauseFn == nil {
		panic("PauseFn is not defined")
	}
	return e.PauseFn()
}

func (e *EngineService) Resume() error {
	if e.ResumeFn == nil {
		panic("ResumeFn is not defined")
	}
	return e.ResumeFn()
}

func (e *EngineService) State() entity.VmState {
	if e.StateFn == nil {
		panic("StateFn is not defined")
	}
	return e.StateFn()
}

func (e *EngineService) Stop() error {
	if e.StopFn == nil {
		panic("StopFn is not defined")
	}
	return e.StopFn()
}

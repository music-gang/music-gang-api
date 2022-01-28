package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.LockService = (*LockService)(nil)

type LockService struct {
	LockFn   func(ctx context.Context)
	NameFn   func() string
	UnlockFn func(ctx context.Context)
}

func (ls *LockService) Lock(ctx context.Context) {
	if ls.LockFn == nil {
		panic("Lock is not defined")
	}
	ls.LockFn(ctx)
}

func (ls *LockService) Name() string {
	if ls.NameFn == nil {
		panic("Name is not defined")
	}
	return ls.NameFn()
}

func (ls *LockService) Unlock(ctx context.Context) {
	if ls.UnlockFn == nil {
		panic("Unlock is not defined")
	}
	ls.UnlockFn(ctx)
}

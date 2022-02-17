package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.LockService = (*LockService)(nil)

type LockService struct {
	LockContextFn   func(ctx context.Context) error
	NameFn          func() string
	UnlockContextFn func(ctx context.Context) (bool, error)
}

func (ls *LockService) LockContext(ctx context.Context) error {
	if ls.LockContextFn == nil {
		panic("Lock is not defined")
	}
	return ls.LockContextFn(ctx)
}

func (ls *LockService) Name() string {
	if ls.NameFn == nil {
		panic("Name is not defined")
	}
	return ls.NameFn()
}

func (ls *LockService) UnlockContext(ctx context.Context) (bool, error) {
	if ls.UnlockContextFn == nil {
		panic("Unlock is not defined")
	}
	return ls.UnlockContextFn(ctx)
}

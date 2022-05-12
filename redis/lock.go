package redis

import (
	"context"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/music-gang/music-gang-api/app/apperr"
)

// LockService implements a distributed lock.
type LockService struct {
	db   *DB
	name string

	mux *redsync.Mutex
}

// NewLockService creates a new LockService.
func NewLockService(db *DB, name string) *LockService {
	pool := goredis.NewPool(db.client)
	rs := redsync.New(pool)
	return &LockService{
		db:   db,
		name: name,
		mux:  rs.NewMutex(name),
	}
}

// LockContext locks the lock.
func (l *LockService) LockContext(ctx context.Context) error {
	if err := l.mux.LockContext(ctx); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to acquire lock: %s", err)
	}
	return nil
}

// Name returns the name of the lock.
func (l *LockService) Name() string {
	return l.name
}

// UnlockContext unlocks the lock.
func (l *LockService) UnlockContext(ctx context.Context) (bool, error) {
	ok, err := l.mux.UnlockContext(ctx)
	if err != nil {
		return false, apperr.Errorf(apperr.EINTERNAL, "failed to release lock: %s", err)
	}
	return ok, nil
}

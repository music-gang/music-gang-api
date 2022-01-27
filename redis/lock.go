package redis

import (
	"context"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
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

// Lock locks the lock.
func (l *LockService) Lock(ctx context.Context) {
	l.mux.LockContext(ctx)
}

// Name returns the name of the lock.
func (l *LockService) Name() string {
	return l.name
}

// Unlock unlocks the lock.
func (l *LockService) Unlock(ctx context.Context) {
	l.mux.UnlockContext(ctx)
}

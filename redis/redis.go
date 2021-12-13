package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/music-gang/music-gang-api/app/apperr"
)

// DB is a wrapper for redis.Client.
type DB struct {
	client *redis.Client
	ctx    context.Context
	cancel func()

	Addr     string
	Password string
}

// NewDB creates a new redis connection.
func NewDB(Addr, Password string) *DB {
	db := &DB{
		Addr:     Addr,
		Password: Password,
	}

	db.ctx, db.cancel = context.WithCancel(context.Background())

	return db
}

// Close closes the redis connection.
func (db *DB) Close() error {
	db.cancel()
	if db.client != nil {
		return db.client.Close()
	}
	return nil
}

// Open opens a new redis connection.
func (db *DB) Open() error {

	db.client = redis.NewClient(&redis.Options{
		Addr:     db.Addr,
		Password: db.Password,
	})

	if res := db.client.Ping(db.ctx); res.Err() != nil {
		return apperr.Errorf(apperr.EINTERNAL, "redis ping error: %s", res.Err())
	}

	return nil
}

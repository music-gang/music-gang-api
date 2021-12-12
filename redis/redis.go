package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/music-gang/music-gang-api/app/apperr"
)

type DB struct {
	client *redis.Client
	ctx    context.Context
	cancel func()

	Addr     string
	Password string
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

	db.ctx, db.cancel = context.WithCancel(context.Background())

	if res := db.client.Ping(db.ctx); res.Err() != nil {
		return apperr.Errorf(apperr.EINTERNAL, "redis ping error: %s", res.Err())
	}

	return nil
}

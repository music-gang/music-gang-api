package redis

import "context"

func (db *DB) FlushAll(ctx context.Context) error {
	return db.client.FlushAll(ctx).Err()
}

func (db *DB) Cancel() {
	db.cancel()
}

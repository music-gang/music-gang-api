package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/redis"
)

func TestLock_lock(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		muxName := "test"

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		lockService := redis.NewLockService(db, muxName)

		lockService.Lock(ctx)

		if lockService.Name() != muxName {
			t.Errorf("got %q, want %q", lockService.Name(), muxName)
		}

		lockService.Unlock(ctx)
	})

	t.Run("CanAcquireLock", func(t *testing.T) {

		muxName := "test"

		db1 := MustOpenDB(t)
		defer db1.Close()

		db2 := MustOpenDB(t)
		defer db2.Close()

		ctx1 := context.Background()
		ctx2 := context.Background()

		if err := db1.FlushAll(ctx1); err != nil {
			t.Fatal(err)
		}

		lockService1 := redis.NewLockService(db1, muxName)
		lockService2 := redis.NewLockService(db2, muxName)

		lockService1.Lock(ctx1)
		lockService1.Unlock(ctx1)

		lockService2.Lock(ctx2)
		lockService2.Unlock(ctx2)
	})

	t.Run("CannotAcquireLock", func(t *testing.T) {

		muxName := "test"

		db1 := MustOpenDB(t)
		defer db1.Close()

		db2 := MustOpenDB(t)
		defer db2.Close()

		ctx1 := context.Background()
		ctx2, cancel2 := context.WithCancel(context.Background())

		if err := db1.FlushAll(ctx1); err != nil {
			t.Fatal(err)
		}

		lockService1 := redis.NewLockService(db1, muxName)
		lockService2 := redis.NewLockService(db2, muxName)

		go func() {
			time.Sleep(1 * time.Second)
			cancel2()
		}()

		lockService1.Lock(ctx1)

		lockService2.Lock(ctx2)

		lockService1.Unlock(ctx1)
	})
}

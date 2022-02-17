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

		if err := lockService.LockContext(ctx); err != nil {
			t.Fatal(err)
		}

		if lockService.Name() != muxName {
			t.Errorf("got %q, want %q", lockService.Name(), muxName)
		}

		if ok, err := lockService.UnlockContext(ctx); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Error("got false, want true")
		}
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

		if err := lockService1.LockContext(ctx1); err != nil {
			t.Fatal(err)
		}
		if ok, err := lockService1.UnlockContext(ctx1); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Error("got false, want true")
		}
		if err := lockService2.LockContext(ctx2); err != nil {
			t.Fatal(err)
		}
		if ok, err := lockService2.UnlockContext(ctx2); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Error("got false, want true")
		}
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

		if err := lockService1.LockContext(ctx1); err != nil {
			t.Fatal(err)
		}

		if err := lockService2.LockContext(ctx2); err == nil {
			t.Error("got nil, want error")
		}

		if ok, err := lockService1.UnlockContext(ctx1); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Error("got false, want true")
		}
	})
}

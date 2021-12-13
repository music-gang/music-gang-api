package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/redis"
)

func TestJWTBlacklist_Invalidate(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		jwtBlacklist := redis.NewJWTBlacklistService(db)

		if err := jwtBlacklist.Invalidate(ctx, "token", time.Second*10); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("InvalidateError", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		jwtBlacklist := redis.NewJWTBlacklistService(db)

		ctx, cancel := context.WithCancel(ctx)

		cancel()

		if err := jwtBlacklist.Invalidate(ctx, "token", time.Second*10); err == nil {
			t.Fatal("Expected error")
		}
	})

	t.Run("InvalidateExpired", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		jwtBlacklist := redis.NewJWTBlacklistService(db)

		if err := jwtBlacklist.Invalidate(ctx, "token", time.Second*1); err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second * 2)

		if isBlacklisted, err := jwtBlacklist.IsBlacklisted(ctx, "token"); err != nil {
			t.Fatal(err)
		} else if isBlacklisted {
			t.Fatal("Expected not blacklisted")
		}
	})

	t.Run("InvalidateNotExpired", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		jwtBlacklist := redis.NewJWTBlacklistService(db)

		if err := jwtBlacklist.Invalidate(ctx, "token", time.Second*1); err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Millisecond * 10)

		if isBlacklisted, err := jwtBlacklist.IsBlacklisted(ctx, "token"); err != nil {
			t.Fatal(err)
		} else if !isBlacklisted {
			t.Fatal("Expected blacklisted")
		}
	})
}

func TestJWTBlacklist_IsBlacklisted(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		jwtBlacklist := redis.NewJWTBlacklistService(db)

		if err := jwtBlacklist.Invalidate(ctx, "token", time.Second*10); err != nil {
			t.Fatal(err)
		}

		if isBlacklisted, err := jwtBlacklist.IsBlacklisted(ctx, "token"); err != nil {
			t.Fatal(err)
		} else if !isBlacklisted {
			t.Fatal("Expected blacklisted")
		}
	})

	t.Run("IsNotBlacklisted", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		jwtBlacklist := redis.NewJWTBlacklistService(db)

		if err := jwtBlacklist.Invalidate(ctx, "token", time.Second*10); err != nil {
			t.Fatal(err)
		}

		if isBlacklisted, err := jwtBlacklist.IsBlacklisted(ctx, "not_this_token"); err != nil {
			t.Fatal(err)
		} else if isBlacklisted {
			t.Fatal("Expected not blacklisted")
		}
	})

	t.Run("IsBlacklistedError", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		jwtBlacklist := redis.NewJWTBlacklistService(db)

		if err := jwtBlacklist.Invalidate(ctx, "token", time.Second*10); err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithCancel(ctx)

		cancel()

		if _, err := jwtBlacklist.IsBlacklisted(ctx, "token"); err == nil {
			t.Fatal("Expected error")
		}
	})
}

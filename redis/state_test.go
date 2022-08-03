package redis_test

import (
	"context"
	"testing"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/redis"
)

func TestState_Cache(t *testing.T) {

	user := &entity.User{
		ID:   1,
		Name: "test-state-cache",
	}

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := app.NewContextWithUser(context.Background(), user)

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		stateService := redis.NewStateService(db)

		state := &entity.State{
			ID:         1,
			RevisionID: 1,
			UserID:     user.ID,
			Value: entity.StateValue{
				"test": "test",
			},
		}

		if err := stateService.CacheState(ctx, state); err != nil {
			t.Fatal(err)
		}

		if s, err := stateService.FindStateByRevisionID(ctx, state.RevisionID); err != nil {
			t.Fatal(err)
		} else if s.Value["test"] != "test" {
			t.Errorf("got %s, want %s", s.Value["test"], "test")
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		stateService := redis.NewStateService(db)

		state := &entity.State{
			ID:         1,
			RevisionID: 1,
			UserID:     user.ID,
			Value: entity.StateValue{
				"test": "test",
			},
		}

		if err := stateService.CacheState(ctx, state); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EUNAUTHORIZED {
			t.Errorf("got %v, want %v", errCode, apperr.EUNAUTHORIZED)
		}
	})

	t.Run("InvalidRevision", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := app.NewContextWithUser(context.Background(), user)

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		stateService := redis.NewStateService(db)

		state := &entity.State{
			ID:     1,
			UserID: user.ID,
			Value: entity.StateValue{
				"test": "test",
			},
		}

		if err := stateService.CacheState(ctx, state); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINVALID {
			t.Errorf("got %v, want %v", errCode, apperr.EINVALID)
		}
	})

	t.Run("ContextCanceled", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := app.NewContextWithUser(context.Background(), user)

		ctx, cancel := context.WithCancel(ctx)
		cancel()

		stateService := redis.NewStateService(db)

		state := &entity.State{
			ID:         1,
			RevisionID: 1,
			UserID:     user.ID,
			Value: entity.StateValue{
				"test": "test",
			},
		}

		if err := stateService.CacheState(ctx, state); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Errorf("got %v, want %v", errCode, apperr.EINTERNAL)
		}
	})
}

func TestState_FindStateByRevisionID(t *testing.T) {

	user := &entity.User{
		ID:   1,
		Name: "test-state-find-by-revision-id",
	}

	state := &entity.State{
		ID:         1,
		RevisionID: 1,
		UserID:     user.ID,
		Value: entity.StateValue{
			"test": "test",
		},
	}

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := app.NewContextWithUser(context.Background(), user)

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		stateService := redis.NewStateService(db)

		MustCacheState(t, ctx, db, state)

		if s, err := stateService.FindStateByRevisionID(ctx, state.RevisionID); err != nil {
			t.Fatal(err)
		} else if s.Value["test"] != "test" {
			t.Errorf("got %s, want %s", s.Value["test"], "test")
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		stateService := redis.NewStateService(db)

		MustCacheState(t, app.NewContextWithUser(ctx, user), db, state)

		if _, err := stateService.FindStateByRevisionID(ctx, state.RevisionID); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EUNAUTHORIZED {
			t.Errorf("got %v, want %v", errCode, apperr.EUNAUTHORIZED)
		}
	})

	t.Run("InvalidRevision", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := app.NewContextWithUser(context.Background(), user)

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		stateService := redis.NewStateService(db)

		MustCacheState(t, app.NewContextWithUser(ctx, user), db, state)

		if _, err := stateService.FindStateByRevisionID(ctx, 0); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINVALID {
			t.Errorf("got %v, want %v", errCode, apperr.EINVALID)
		}
	})

	t.Run("NotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := app.NewContextWithUser(context.Background(), user)

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		stateService := redis.NewStateService(db)

		if _, err := stateService.FindStateByRevisionID(ctx, 1); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.ENOTFOUND {
			t.Errorf("got %v, want %v", errCode, apperr.ENOTFOUND)
		}
	})

	t.Run("ContextCanceled", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := app.NewContextWithUser(context.Background(), user)

		ctx, cancel := context.WithCancel(ctx)

		stateService := redis.NewStateService(db)

		MustCacheState(t, ctx, db, state)
		cancel()

		if _, err := stateService.FindStateByRevisionID(ctx, state.RevisionID); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Errorf("got %v, want %v", errCode, apperr.EINTERNAL)
		}
	})
}

func MustCacheState(t testing.TB, ctx context.Context, db *redis.DB, state *entity.State) {

	stateService := redis.NewStateService(db)

	if err := stateService.CacheState(ctx, state); err != nil {
		t.Fatal(err)
	}
}

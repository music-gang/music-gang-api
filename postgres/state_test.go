package postgres_test

import (
	"context"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/mock"
	"github.com/music-gang/music-gang-api/postgres"
)

func TestStateService_CreateState(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		contract := &entity.Contract{
			Name:       "test",
			MaxFuel:    entity.FuelInstantActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		rev, ctx := MustCreateRevision(t, context.Background(), db, DataToMakeRevision{
			Contract: contract,
			User:     &entity.User{Name: "test-create-state"},
			Revision: &entity.Revision{
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
		})

		state := &entity.State{
			RevisionID: rev.ID,
			Value:      make(entity.StateValue),
		}

		if err := s.CreateState(ctx, state); err != nil {
			t.Fatal("unexpected error:", err)
		}

		if state.ID == 0 {
			t.Fatal("state ID is 0")
		}
	})

	t.Run("Invalid", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		user := &entity.User{Name: "test-create-state-invalid"}

		contract := &entity.Contract{
			Name:       "test",
			MaxFuel:    entity.FuelInstantActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		rev, ctx := MustCreateRevision(t, context.Background(), db, DataToMakeRevision{
			Contract: contract,
			User:     user,
			Revision: &entity.Revision{
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
		})

		invalidStates := map[string]*entity.State{
			"InvalidRevisionID": {
				RevisionID: 0,
				UserID:     user.ID,
				Value:      make(entity.StateValue),
			},
			"InvalidValue": {
				RevisionID: rev.ID,
				UserID:     user.ID,
				Value:      nil,
			},
		}

		for key, state := range invalidStates {
			t.Run(key, func(t *testing.T) {

				if err := s.CreateState(ctx, state); err == nil {
					t.Fatal("expected error")
				} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINVALID {
					t.Fatalf("expected %s, got %s", apperr.EINVALID, errCode)
				}
			})
		}
	})

	t.Run("ContextCacelled", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if err := s.CreateState(ctx, &entity.State{}); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})

	t.Run("UserNotAuthenticated", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx := context.Background()

		if err := s.CreateState(ctx, &entity.State{}); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EUNAUTHORIZED {
			t.Fatalf("expected %s, got %s", apperr.EUNAUTHORIZED, errCode)
		}
	})

	t.Run("ErrCreateLockService", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return nil, apperr.Errorf(apperr.EINTERNAL, "test")
		}

		ctx := context.Background()

		if err := s.CreateState(ctx, &entity.State{}); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})

	t.Run("CannotAcquireLock", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return apperr.Errorf(apperr.EINTERNAL, "test")
				},
			}, nil
		}

		ctx := context.Background()

		if err := s.CreateState(ctx, &entity.State{}); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})
}

func TestStateService_FindStateByRevisionID(t *testing.T) {
	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx := context.Background()

		state, ctx := MustCreateState(t, ctx, db, DataToMakeState{
			User: &entity.User{Name: "test-find-state-by-revision-id"},
			Contract: &entity.Contract{
				Name:       "test",
				MaxFuel:    entity.FuelInstantActionAmount,
				Visibility: entity.VisibilityPublic,
			},
			Revision: &entity.Revision{
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
			State: &entity.State{
				Value: make(entity.StateValue),
			},
		})

		if stateFetched, err := s.FindStateByRevisionID(ctx, state.RevisionID); err != nil {
			t.Fatal("unexpected error:", err)
		} else if stateFetched.ID != state.ID {
			t.Fatal("state IDs do not match")
		}
	})

	t.Run("FromCache", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		s.CacheStateSearchService = &mock.StateService{
			FindStateByRevisionIDFn: func(ctx context.Context, revisionID int64) (*entity.State, error) {
				return &entity.State{ID: 1, RevisionID: 1, UserID: 1, Value: make(entity.StateValue)}, nil
			},
		}

		ctx := context.Background()

		if state, err := s.FindStateByRevisionID(ctx, 1); err != nil {
			t.Fatal("unexpected error:", err)
		} else if state.ID != 1 {
			t.Fatal("state IDs do not match")
		}
	})

	t.Run("ContextCacelled", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := s.FindStateByRevisionID(ctx, 1); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})

	t.Run("UserNotAuthenticated", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx := context.Background()

		if _, err := s.FindStateByRevisionID(ctx, 1); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EUNAUTHORIZED {
			t.Fatalf("expected %s, got %s", apperr.EUNAUTHORIZED, errCode)
		}
	})

	t.Run("NotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx := context.Background()

		_, ctx = MustCreateState(t, ctx, db, DataToMakeState{
			User: &entity.User{Name: "test-find-state-by-revision-id"},
			Contract: &entity.Contract{
				Name:       "test",
				MaxFuel:    entity.FuelInstantActionAmount,
				Visibility: entity.VisibilityPublic,
			},
			Revision: &entity.Revision{
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
			State: &entity.State{
				Value: make(entity.StateValue),
			},
		})

		if _, err := s.FindStateByRevisionID(ctx, 2); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, errCode)
		}
	})

	t.Run("ErrCreateLockService", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return nil, apperr.Errorf(apperr.EINTERNAL, "test")
		}

		ctx := context.Background()

		if _, err := s.FindStateByRevisionID(ctx, 1); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})

	t.Run("CannotAcquireLock", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return apperr.Errorf(apperr.EINTERNAL, "test")
				},
			}, nil
		}

		ctx := context.Background()

		if _, err := s.FindStateByRevisionID(ctx, 1); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})
}

func TestStateService_UpdateState(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx := context.Background()

		state, ctx := MustCreateState(t, ctx, db, DataToMakeState{
			User: &entity.User{Name: "test-update-state"},
			Contract: &entity.Contract{
				Name:       "test",
				MaxFuel:    entity.FuelInstantActionAmount,
				Visibility: entity.VisibilityPublic,
			},
			Revision: &entity.Revision{
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
			State: &entity.State{
				Value: make(entity.StateValue),
			},
		})

		if ss, err := s.UpdateState(ctx, state.RevisionID, entity.StateValue{
			"test": "test",
		}); err != nil {
			t.Fatal("unexpected error:", err)
		} else if ss.ID != state.ID {
			t.Fatal("state IDs do not match")
		} else if v, ok := ss.Value["test"]; !ok {
			t.Fatal("state value does not contain key")
		} else if v != "test" {
			t.Fatal("state value does not contain correct value")
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx := context.Background()

		state, ctx := MustCreateState(t, ctx, db, DataToMakeState{
			User: &entity.User{Name: "test-update-state"},
			Contract: &entity.Contract{
				Name:       "test",
				MaxFuel:    entity.FuelInstantActionAmount,
				Visibility: entity.VisibilityPublic,
			},
			Revision: &entity.Revision{
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
			State: &entity.State{
				Value: make(entity.StateValue),
			},
		})

		if _, err := s.UpdateState(ctx, state.RevisionID, nil); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINVALID {
			t.Fatalf("expected %s, got %s", apperr.EINVALID, errCode)
		}
	})

	t.Run("ContextCacelled", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := s.UpdateState(ctx, 1, entity.StateValue{}); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})

	t.Run("UserNotAuthenticated", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx := context.Background()

		if _, err := s.UpdateState(ctx, 1, entity.StateValue{}); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EUNAUTHORIZED {
			t.Fatalf("expected %s, got %s", apperr.EUNAUTHORIZED, errCode)
		}
	})

	t.Run("NotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return nil
				},
				UnlockContextFn: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			}, nil
		}

		ctx := context.Background()

		_, ctx = MustCreateUser(t, ctx, db, &entity.User{
			Name: "test-update-state-not-found",
		})

		if _, err := s.UpdateState(ctx, 1, entity.StateValue{}); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, errCode)
		}
	})

	t.Run("ErrCreateLockService", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return nil, apperr.Errorf(apperr.EINTERNAL, "test")
		}

		ctx := context.Background()

		_, ctx = MustCreateUser(t, ctx, db, &entity.User{
			Name: "test-update-state-not-found",
		})

		if _, err := s.UpdateState(ctx, 1, entity.StateValue{}); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})

	t.Run("CannotAcquireLock", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTableForStateTests(t, db)

		s := postgres.NewStateService(db)

		s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
			return &mock.LockService{
				LockContextFn: func(ctx context.Context) error {
					return apperr.Errorf(apperr.EINTERNAL, "test")
				},
			}, nil
		}

		ctx := context.Background()

		_, ctx = MustCreateUser(t, ctx, db, &entity.User{
			Name: "test-update-state-not-found",
		})

		if _, err := s.UpdateState(ctx, 1, entity.StateValue{}); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})
}

type DataToMakeState struct {
	User     *entity.User
	Contract *entity.Contract
	Revision *entity.Revision
	State    *entity.State
}

func MustCreateState(t testing.TB, ctx context.Context, db *postgres.DB, data DataToMakeState) (*entity.State, context.Context) {
	s := postgres.NewStateService(db)

	s.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
		return &mock.LockService{
			LockContextFn: func(ctx context.Context) error {
				return nil
			},
			UnlockContextFn: func(ctx context.Context) (bool, error) {
				return true, nil
			},
		}, nil
	}

	rev, ctx := MustCreateRevision(t, ctx, db, DataToMakeRevision{
		Contract: data.Contract,
		User:     data.User,
		Revision: data.Revision,
	})

	data.State.RevisionID = rev.ID

	if err := s.CreateState(ctx, data.State); err != nil {
		t.Fatal("unexpected error:", err)
	}

	return data.State, ctx
}

func MustTruncateTableForStateTests(t testing.TB, db *postgres.DB) {
	MustTruncateTable(t, db, "users")
	MustTruncateTable(t, db, "contracts")
	MustTruncateTable(t, db, "revisions")
	MustTruncateTable(t, db, "states")
}

package postgres

import (
	"context"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/postgres/query"
)

var _ service.StateService = (*StateService)(nil)

// StateService is the implementation of service.StateService for PostgreSQL.
type StateService struct {
	db *DB

	// CacheStateSearchService is the cache for searching states.
	// It is used to avoid querying the database when the state is already in the cache.
	// Can be nil if the cache is not enabled.
	CacheStateSearchService service.StateSearchService

	// LockService is the service for locking the state during I/O operations.
	CreateLockService func(ctx context.Context, revisionID int64) (service.LockService, error)
}

// NewStateService creates a new StateService.
func NewStateService(db *DB) *StateService {
	return &StateService{
		db: db,
	}
}

// CreateState creates a new state.
func (s *StateService) CreateState(ctx context.Context, state *entity.State) error {

	ls, err := s.CreateLockService(ctx, state.RevisionID)
	if err != nil {
		return err
	}
	if err := ls.LockContext(ctx); err != nil {
		return err
	}
	defer ls.UnlockContext(ctx)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createState(ctx, tx, state); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return nil
}

// FindStateByRevisionID finds the state by revision ID and the authenticated user retrieved from the context.
// If cache is enabled, it tries to find the state in the cache first.
func (s *StateService) FindStateByRevisionID(ctx context.Context, revisionID int64) (*entity.State, error) {

	ls, err := s.CreateLockService(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	if err := ls.LockContext(ctx); err != nil {
		return nil, err
	}
	defer ls.UnlockContext(ctx)

	if s.CacheStateSearchService != nil {
		state, err := s.CacheStateSearchService.FindStateByRevisionID(ctx, revisionID)
		if err == nil {
			return state, nil
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	state, err := findStateByRevisionID(ctx, tx, revisionID)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// UpdateState updates the state.
func (s *StateService) UpdateState(ctx context.Context, revisionID int64, value entity.StateValue) (*entity.State, error) {

	ls, err := s.CreateLockService(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	if err := ls.LockContext(ctx); err != nil {
		return nil, err
	}
	defer ls.UnlockContext(ctx)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	state, err := updateState(ctx, tx, revisionID, value)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return state, nil
}

// createState creates a new state.
func createState(ctx context.Context, tx *Tx, state *entity.State) error {

	userID := app.UserIDFromContext(ctx)
	if userID == 0 {
		return apperr.Errorf(apperr.EUNAUTHORIZED, "user is not authorized to create the state")
	}

	state.UserID = userID
	state.CreatedAt = tx.now
	state.UpdatedAt = tx.now

	if err := state.Validate(); err != nil {
		return err
	}

	if err := tx.QueryRowContext(ctx, query.InsertStateQuery(),
		state.RevisionID,
		state.Value,
		state.UserID,
		state.CreatedAt,
		state.UpdatedAt).Scan(&state.ID); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to insert state: %v", err)
	}

	return nil
}

// findStateByRevisionID finds the state by revision ID and the authenticated user retrieved from the context.
func findStateByRevisionID(ctx context.Context, tx *Tx, revisionID int64) (*entity.State, error) {

	userID := app.UserIDFromContext(ctx)
	if userID == 0 {
		return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "user is not authorized to access the state")
	}

	rows, err := tx.QueryContext(ctx, query.SelectStateByRevisionIDAndUserIDQuery(), revisionID, userID)
	if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to select state: %v", err)
	}
	defer rows.Close()

	var state entity.State

	if rows.Next() {

		if err := rows.Scan(
			&state.ID,
			&state.RevisionID,
			&state.Value,
			&state.UserID,
			&state.CreatedAt,
			&state.UpdatedAt); err != nil {
			return nil, apperr.Errorf(apperr.EINTERNAL, "failed to scan state: %v", err)
		}

	} else {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "state not found")
	}

	return &state, nil
}

// updateState updates the state.
func updateState(ctx context.Context, tx *Tx, revisionID int64, value entity.StateValue) (*entity.State, error) {

	userID := app.UserIDFromContext(ctx)
	if userID == 0 {
		return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "user is not authorized to update the state")
	}

	state, err := findStateByRevisionID(ctx, tx, revisionID)
	if err != nil {
		return nil, err
	}

	if state.UserID != userID {
		return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "user is not the owner of the state")
	}

	state.Value = value
	state.UpdatedAt = tx.now

	if err := state.Validate(); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, query.UpdateStateQuery(),
		state.Value,
		state.UpdatedAt,
		state.ID); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to update state: %v", err)
	}

	return state, nil

}

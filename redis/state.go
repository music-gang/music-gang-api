package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

const (
	StateLockKeyTemplate = "state-user-%d-revision-%d-lock"
	StateKeyTemplate     = "state-user-%d-revision-%d"
)

var StateCachePeriod = 10 * time.Minute

var _ service.StateCacheService = (*StateService)(nil)
var _ service.StateSearchService = (*StateService)(nil)

// StateService implements the StateService for Redis.
type StateService struct {
	db *DB
}

// NewStateService creates a new StateService.
func NewStateService(db *DB) *StateService {
	return &StateService{db: db}
}

// CacheState caches a state.
func (s *StateService) CacheState(ctx context.Context, state *entity.State) error {
	return cacheState(ctx, s.db, state)
}

// FindStateByRevisionID finds a state by revision ID and user injected into the context.
func (s *StateService) FindStateByRevisionID(ctx context.Context, revisionID int64) (*entity.State, error) {
	return findStateByRevisionID(ctx, s.db, revisionID)
}

// cacheState caches a state.
// Cache period is 10 minutes.
func cacheState(ctx context.Context, db *DB, state *entity.State) error {

	userID := app.UserIDFromContext(ctx)
	if userID == 0 {
		return apperr.Errorf(apperr.EUNAUTHORIZED, "user not authorized")
	}

	if state.RevisionID == 0 {
		return apperr.Errorf(apperr.EINVALID, "revisionID is 0")
	}

	key := fmt.Sprintf(StateKeyTemplate, userID, state.RevisionID)

	rawVal, err := json.Marshal(state)
	if err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to marshal state: %v", err)
	}

	if err := db.client.Set(ctx, key, string(rawVal), StateCachePeriod).Err(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to set state to redis: %v", err)
	}

	return nil
}

// findStateByRevisionID finds a state by revision ID and user injected into the context.
// Return ENOTFOUND if no state is found.
func findStateByRevisionID(ctx context.Context, db *DB, revisionID int64) (*entity.State, error) {

	userID := app.UserIDFromContext(ctx)
	if userID == 0 {
		return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "user not authorized")
	}

	if revisionID == 0 {
		return nil, apperr.Errorf(apperr.EINVALID, "revisionID is 0")
	}

	key := fmt.Sprintf(StateKeyTemplate, userID, revisionID)

	rawVal, err := db.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "state not found")
	} else if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to get state from redis: %v", err)
	}

	var state entity.State

	if err := json.Unmarshal([]byte(rawVal), &state); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to unmarshal state: %v", err)
	}

	return &state, nil
}

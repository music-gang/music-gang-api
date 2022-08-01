package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.StateService = (*StateService)(nil)

type StateService struct {
	FindStateByRevisionIDFn func(ctx context.Context, revisionID int64) (*entity.State, error)
	CreateStateFn           func(ctx context.Context, state *entity.State) error
	UpdateStateFn           func(ctx context.Context, revisionID int64, value entity.StateValue) (*entity.State, error)
}

func (s *StateService) FindStateByRevisionID(ctx context.Context, revisionID int64) (*entity.State, error) {
	if s.FindStateByRevisionIDFn == nil {
		panic("FindStateByRevisionID not defined")
	}
	return s.FindStateByRevisionIDFn(ctx, revisionID)
}

func (s *StateService) CreateState(ctx context.Context, state *entity.State) error {
	if s.CreateStateFn == nil {
		panic("CreateState not defined")
	}
	return s.CreateStateFn(ctx, state)
}

func (s *StateService) UpdateState(ctx context.Context, revisionID int64, value entity.StateValue) (*entity.State, error) {
	if s.UpdateStateFn == nil {
		panic("UpdateState not defined")
	}
	return s.UpdateStateFn(ctx, revisionID, value)
}

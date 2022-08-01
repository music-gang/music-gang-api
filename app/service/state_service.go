package service

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
)

// StateSearchService is the interface for searching states.
type StateSearchService interface {
	// FindStateByRevisionID finds the state by revision ID and the authenticated user retrieved from the context.
	// Should returns ENOTFOUND if the state is not found.
	// Should returns EUNAUTHORIZED if the user is not authorized to access the state.
	FindStateByRevisionID(ctx context.Context, revisionID int64) (*entity.State, error)
}

// StateManagementService is the interface for managing states.
type StateManagementService interface {
	// CreateState creates a state.
	// Should returns EUNAUTHORIZED if the user is not authorized to create the state.
	CreateState(ctx context.Context, state *entity.State) error
	// UpdateState updates a state.
	// Should returns ENOTFOUND if the state is not found.
	// Should returns EUNAUTHORIZED if the user is not authorized to access the state.
	UpdateState(ctx context.Context, revisionID int64, value entity.StateValue) (*entity.State, error)
}

// StateService is the interface for managing and searching states.
type StateService interface {
	StateSearchService
	StateManagementService
}

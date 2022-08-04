package handler

import (
	"context"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

// CurrentAuthUser returns the current authenticated user from the passed context.
func (s *ServiceHandler) CurrentAuthUser(ctx context.Context) (*entity.User, error) {
	user, err := app.AuthUser(ctx)
	if err != nil {
		s.LogService.ReportFatal(ctx, err)
		return nil, err
	}

	return user, nil
}

// UpdateUser Updates the user with the given ID.
func (s *ServiceHandler) UpdateUser(ctx context.Context, userID int64, userParams service.UserUpdate) (*entity.User, error) {

	user, err := s.UserSearchService.FindUserByID(ctx, userID)
	if err != nil {
		s.LogService.ReportFatal(ctx, err)
		return nil, err
	}

	if updatedUser, err := s.VmCallableService.UpdateUser(ctx, user.ID, userParams); err != nil {
		s.LogService.ReportError(ctx, err)
		return nil, err
	} else {
		return updatedUser, nil
	}
}

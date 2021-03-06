package postgres

import (
	"context"
	"fmt"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/postgres/query"
	"gopkg.in/guregu/null.v4"
)

var _ service.AuthService = (*AuthService)(nil)

// AuthService is a service for managing authentication.
type AuthService struct {
	db *DB
}

// NewAuthService creates a new AuthService.
func NewAuthService(db *DB) *AuthService {
	return &AuthService{db}
}

// Auhenticate not implemented.
func (a *AuthService) Auhenticate(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {
	return nil, apperr.Errorf(apperr.ENOTIMPLEMENTED, "not implemented")
}

// CreateAuth creates a new auth.
// If is attached to a user, links the auth to the user, otherwise creates a new user.
// On success, the auth.ID is set.
func (s *AuthService) CreateAuth(ctx context.Context, auth *entity.Auth) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if auth.SourceID.Valid {
		// Check to see if the auth already exists for the given source.
		other, err := findAuthBySourceID(ctx, tx, auth.Source, auth.SourceID.String)
		if err == nil {
			// If an auth already exists for the source user, update with the new tokens.
			other, err := updateAuth(ctx, tx, other.ID, auth.AccessToken, auth.RefreshToken, auth.Expiry)
			if err != nil {
				return err
			}

			if err := attachAuthAssociations(ctx, tx, other); err != nil {
				return err
			}

			// Copy found auth back to the caller's arg & return.
			*auth = *other

			if err := tx.Commit(); err != nil {
				return apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
			}

			return nil

		} else if apperr.ErrorCode(err) != apperr.ENOTFOUND {
			// Check if no auth exists, if err is not ENOTFOUND, than returns err.
			return err
		}
	}

	// check if user had new object passed in. It is considered "new" if the user ID is not set.
	if auth.UserID == 0 && auth.User != nil {

		// new user from an auth source because user ID is not set but auth have attached a user object.
		user, err := findUserByEmail(ctx, tx, auth.User.Email.String)
		if err == nil {
			if !auth.SourceID.Valid {
				return apperr.Errorf(apperr.EFORBIDDEN, "email already exists")
			}
			auth.User = user
		} else if apperr.ErrorCode(err) == apperr.ENOTFOUND {
			if err := createUser(ctx, tx, auth.User); err != nil {
				return err
			}
		} else {
			return err
		}

		auth.UserID = auth.User.ID
	}

	if err := createAuth(ctx, tx, auth); err != nil {
		return err
	} else if err := attachAuthAssociations(ctx, tx, auth); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return nil
}

// DeleteAuth deletes an auth.
// Do not delete underlying user.
func (s *AuthService) DeleteAuth(ctx context.Context, id int64) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteAuth(ctx, tx, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return nil
}

// FindAuthByID returns a single auth by its id.
// Returns ENOTFOUND if the auth does not exist.
func (s *AuthService) FindAuthByID(ctx context.Context, id int64) (*entity.Auth, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	auth, err := findAuthByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := attachAuthAssociations(ctx, tx, auth); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return auth, nil
}

// FindAuths returns a list of auths.
// Predicate can be used to filter the results.
// Also returns the total count of auths.
func (s *AuthService) FindAuths(ctx context.Context, filter service.AuthFilter) (entity.Auths, int, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	auths, total, err := findAuths(ctx, tx, filter)
	if err != nil {
		return nil, 0, err
	}

	return auths, total, nil
}

// attachAuthAssociations attaches user associations to the passed auth.
func attachAuthAssociations(ctx context.Context, tx *Tx, auth *entity.Auth) (err error) {
	if auth.User, err = findUserByID(ctx, tx, auth.UserID); err != nil {
		return err
	}
	return nil
}

// createAuth creates a new auth.
func createAuth(ctx context.Context, tx *Tx, auth *entity.Auth) error {

	auth.CreatedAt = tx.now
	auth.UpdatedAt = tx.now

	if err := auth.Validate(); err != nil {
		return err
	}

	if err := tx.QueryRowContext(ctx, query.InsertAuthQuery(),
		auth.UserID,
		auth.Source,
		auth.SourceID,
		auth.AccessToken,
		auth.RefreshToken,
		auth.Expiry,
		auth.CreatedAt,
		auth.UpdatedAt).Scan(&auth.ID); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to create auth: %v", err)
	}

	return nil
}

// deleteAuth deletes an auth.
// Do not delete underlying user.
// Returns EUNAUTHORIZED if current user is not allowed to delete this auth
// Return EFORBIDDEN if the specified auth cannot be deleted.
func deleteAuth(ctx context.Context, tx *Tx, id int64) error {

	if auth, err := findAuthByID(ctx, tx, id); err != nil {
		return err
	} else if auth.UserID != app.UserIDFromContext(ctx) {
		return apperr.Errorf(apperr.EUNAUTHORIZED, "you are not allowed to delete another user auth")
	} else if !entity.CanAuthBeDeleted(auth) {
		return apperr.Errorf(apperr.EFORBIDDEN, "cannot delete this auth")
	}

	if _, err := tx.ExecContext(ctx, query.DeleteAuthQuery(), id); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to delete auth: %v", err)
	}

	return nil
}

// findAuthByID returns a single auth by its id.
// Returns ENOTFOUND if the auth does not exist.
func findAuthByID(ctx context.Context, tx *Tx, id int64) (*entity.Auth, error) {

	auths, _, err := findAuths(ctx, tx, service.AuthFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(auths) == 0 {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
	}

	return auths[0], nil
}

// findAuthBySourceID is a helper function to return an auth object by source ID.
// Returns ENOTFOUND if auth doesn't exist.
func findAuthBySourceID(ctx context.Context, tx *Tx, source, sourceID string) (*entity.Auth, error) {

	auths, _, err := findAuths(ctx, tx, service.AuthFilter{Source: &source, SourceID: &sourceID})
	if err != nil {
		return nil, err
	} else if len(auths) == 0 {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
	}

	return auths[0], nil
}

// findAuths returns a list of auths.
// Predicate can be used to filter the results.
// Also returns the total count of auths.
func findAuths(ctx context.Context, tx *Tx, filter service.AuthFilter) (_ entity.Auths, n int, err error) {

	where, args := []string{"1 = 1"}, []interface{}{}
	counParameter := 1

	if v := filter.ID; v != nil {
		where = append(where, fmt.Sprintf("id = $%d", counParameter))
		args = append(args, *v)
		counParameter++
	}
	if v := filter.UserID; v != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", counParameter))
		args = append(args, *v)
		counParameter++
	}
	if v := filter.UserID; v != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", counParameter))
		args = append(args, *v)
		counParameter++
	}
	if v := filter.Source; v != nil {
		where = append(where, fmt.Sprintf("source = $%d", counParameter))
		args = append(args, *v)
		counParameter++
	}
	if v := filter.SourceID; v != nil {
		where = append(where, fmt.Sprintf("source_id = $%d", counParameter))
		args = append(args, *v)
		counParameter++
	}

	rows, err := tx.QueryContext(ctx, query.SelectAuthsQuery(where, filter.Limit, filter.Offset), args...)
	if err != nil {
		return nil, 0, apperr.Errorf(apperr.EINTERNAL, "failed to query auths: %v", err)
	}
	defer rows.Close()

	auths := make(entity.Auths, 0)

	for rows.Next() {
		var auth entity.Auth
		if err := rows.Scan(
			&auth.ID,
			&auth.UserID,
			&auth.Source,
			&auth.SourceID,
			&auth.AccessToken,
			&auth.RefreshToken,
			&auth.Expiry,
			&auth.CreatedAt,
			&auth.UpdatedAt,
			&n,
		); err != nil {
			return nil, 0, apperr.Errorf(apperr.EINTERNAL, "failed to scan auth: %v", err)
		}

		auths = append(auths, &auth)

	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.Errorf(apperr.EINTERNAL, "failed to iterate over auths: %v", err)
	}

	return auths, n, nil

}

// updateAuth updates an auth.
// Returns AUNAUTHORIZED if current user is not allowed to update this auth
func updateAuth(ctx context.Context, tx *Tx, id int64, accessToken, refreshToken null.String, expiry null.Time) (*entity.Auth, error) {

	auth, err := findAuthByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if auth.UserID != app.UserIDFromContext(ctx) {
		return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "you are not allowed to update other user auth")
	}

	auth.AccessToken = accessToken
	auth.RefreshToken = refreshToken
	auth.Expiry = expiry
	auth.UpdatedAt = tx.now

	if err := auth.Validate(); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, query.UpdateAuthQuery(),
		accessToken,
		refreshToken,
		expiry,
		tx.now,
		id); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to update auth: %v", err)
	}

	return auth, nil
}

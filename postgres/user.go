package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

// Ensure service implements interface.
var _ service.UserService = (*UserService)(nil)

// UserService represents a service for managing users.
type UserService struct {
	db *DB
}

// NewUserService creates a new user service.
func NewUserService(db *DB) *UserService {
	return &UserService{db}
}

// CreateUser creates a new user.
func (s *UserService) CreateUser(ctx context.Context, user *entity.User) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createUser(ctx, tx, user); err != nil {
		return err
	} else if err := attachUserAssociations(ctx, tx, user); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return nil
}

// DeleteUser deletes the user with the given id.
// Return EUNAUTHORIZED if the user is not the same as the authenticated user.
// Return ENOTFOUND if the user does not exist.
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteUser(ctx, tx, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return nil
}

// FindUserByID returns the user with the given id.
// Return ENOTFOUND if the user does not exist.
func (s *UserService) FindUserByID(ctx context.Context, id int64) (*entity.User, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := attachUserAssociations(ctx, tx, user); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return user, nil
}

// FindUsers retrieves a list of users by filter. Also returns total count of
// matching users which may differ from returned results if filter.Limit is specified.
func (s *UserService) FindUsers(ctx context.Context, filter service.UserFilter) (entity.Users, int, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	return findUsers(ctx, tx, filter)
}

// UpdateUser updates the given user.
// Return EUNAUTHORIZED if the user is not the same as the authenticated user.
// Return ENOTFOUND if the user does not exist.
func (u *UserService) UpdateUser(ctx context.Context, id int64, upd service.UserUpdate) (*entity.User, error) {

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := updateUser(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	} else if err := attachUserAssociations(ctx, tx, user); err != nil {
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return user, nil
}

// attachUserAssociations attaches OAuth objects associated with the user.
func attachUserAssociations(ctx context.Context, tx *Tx, user *entity.User) (err error) {
	if user.Auths, _, err = findAuths(ctx, tx, service.AuthFilter{UserID: &user.ID}); err != nil {
		return err
	}
	return nil
}

func createUser(ctx context.Context, tx *Tx, user *entity.User) error {

	user.CreatedAt = tx.now
	user.UpdatedAt = user.CreatedAt

	if err := user.Validate(); err != nil {
		return err
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO users (
			name,
			email,
			password,
			created_at,
			updated_at
		) VALUES ( $1, $2, $3, $4, $5 ) RETURNING id
	`, user.Name, user.Email, user.Password, user.CreatedAt, user.UpdatedAt).Scan(&user.ID); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to insert user: %v", err)
	}

	return nil
}

// deleteUser deletes the user with the given id.
// Return EUNAUTHORIZED if the user is not the same as the authenticated user.
func deleteUser(ctx context.Context, tx *Tx, id int64) error {

	if user, err := findUserByID(ctx, tx, id); err != nil {
		return err
	} else if user.ID != app.UserIDFromContext(ctx) {
		return apperr.Errorf(apperr.EUNAUTHORIZED, "you are not allowed to delete this user")
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to delete user: %v", err)
	}

	return nil
}

// findUserByEmail returns the user with the given email.
// Return ENOTFOUND if the user does not exist.
func findUserByEmail(ctx context.Context, tx *Tx, email string) (*entity.User, error) {

	a, _, err := findUsers(ctx, tx, service.UserFilter{Email: &email})
	if err != nil {
		return nil, err
	} else if len(a) == 0 {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
	}

	return a[0], nil
}

// findUserByID returns the user with the given id.
// Return ENOTFOUND if the user does not exist.
func findUserByID(ctx context.Context, tx *Tx, id int64) (*entity.User, error) {

	a, _, err := findUsers(ctx, tx, service.UserFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(a) == 0 {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
	}

	return a[0], nil
}

// findUsers returns a list of users matching a filter. Also returns a count of
// total matching users which may differ if filter.Limit is set.
func findUsers(ctx context.Context, tx *Tx, filter service.UserFilter) (_ entity.Users, n int, err error) {

	where, args := []string{"1 = 1"}, []interface{}{}

	counterParameter := 1

	if v := filter.ID; v != nil {
		where = append(where, fmt.Sprintf("id = $%d", counterParameter))
		args = append(args, *v)
		counterParameter++
	}
	if v := filter.Email; v != nil {
		where = append(where, fmt.Sprintf("email = $%d", counterParameter))
		args = append(args, *v)
		counterParameter++
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    id,
		    name,
		    email,
			password,
		    created_at,
		    updated_at,
		    COUNT(*) OVER() as count
		FROM users
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, 0, apperr.Errorf(apperr.EINTERNAL, "failed to query users: %v", err)
	}
	defer rows.Close()

	users := make(entity.Users, 0)

	for rows.Next() {

		var user entity.User

		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
			&n,
		); err != nil {
			return nil, 0, apperr.Errorf(apperr.EINTERNAL, "failed to scan user: %v", err)
		}

		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.Errorf(apperr.EINTERNAL, "failed to iterate over users: %v", err)
	}

	return users, n, nil
}

// updateUser updates the given user.
// Return EUNAUTHORIZED if the user is not the same as the authenticated user.
func updateUser(ctx context.Context, tx *Tx, id int64, upd service.UserUpdate) (*entity.User, error) {

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if user.ID != app.UserIDFromContext(ctx) {
		return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "you are not allowed to update this user")
	}

	if v := upd.Name; v != nil {
		user.Name = *v
	}

	user.UpdatedAt = tx.now

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users SET
			name = $1,
			updated_at = $2
		WHERE id = $3
	`, user.Name, user.UpdatedAt, id); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to update user: %v", err)
	}

	return user, nil
}

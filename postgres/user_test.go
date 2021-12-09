package postgres

import (
	"context"
	"reflect"
	"testing"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"gopkg.in/guregu/null.v4"
)

func TestUserService_CreateUser(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := NewUserService(db)

		// user with email and password
		u1 := &entity.User{
			Email:    null.StringFrom("test.user@domain.com"),
			Name:     "Test User",
			Password: null.StringFrom("password"),
		}

		if err := s.CreateUser(context.Background(), u1); err != nil {
			t.Fatal(err)
		} else if got, want := u1.ID, int64(1); got != want {
			t.Fatalf("got %d, want %d", got, want)
		} else if u1.CreatedAt.IsZero() {
			t.Fatal("created at is zero")
		} else if u1.UpdatedAt.IsZero() {
			t.Fatal("updated at is zero")
		}

		// Create second user with email.
		u2 := &entity.User{Name: "jane"}
		if err := s.CreateUser(context.Background(), u2); err != nil {
			t.Fatal(err)
		} else if got, want := u2.ID, int64(2); got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		}

		if other, err := s.FindUserByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(u1, other) {
			t.Fatalf("mismatch: %#v != %#v", u1, other)
		}
	})

	t.Run("ErrNameRequired", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := NewUserService(db)

		if err := s.CreateUser(context.Background(), &entity.User{}); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINVALID {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EINVALID)
		}
	})
}

func MustCreateUser(tb testing.TB, ctx context.Context, db *DB, user *entity.User) (*entity.User, context.Context) {
	tb.Helper()
	if err := NewUserService(db).CreateUser(ctx, user); err != nil {
		tb.Fatal(err)
	}
	return user, app.NewContextWithUser(ctx, user)
}

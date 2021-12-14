package postgres_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/postgres"
	"gopkg.in/guregu/null.v4"
)

func TestUserService_CreateUser(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

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
		u2 := &entity.User{Name: "Jane"}
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

		s := postgres.NewUserService(db)

		if err := s.CreateUser(context.Background(), &entity.User{}); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINVALID {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EINVALID)
		}
	})

	t.Run("ErrEmailNotEmpty", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		if err := s.CreateUser(context.Background(), &entity.User{Name: "Jane", Email: null.StringFrom("")}); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINVALID {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EINVALID)
		}
	})

	t.Run("ErrCtxDone", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		ctxToCancel, cancel := context.WithCancel(context.Background())

		cancel()

		if err := s.CreateUser(ctxToCancel, &entity.User{Name: "Jane"}); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINTERNAL {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EINTERNAL)
		}
	})
}

func TestUserService_UpdateUser(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, ctx0 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane"})

		newName := "Jane Doe"

		uu, err := s.UpdateUser(ctx0, user0.ID, service.UserUpdate{Name: &newName})
		if err != nil {
			t.Fatal(err)
		} else if got, want := uu.Name, newName; got != want {
			t.Fatalf("got %q, want %q", got, want)
		} else if got, want := uu.UpdatedAt, user0.UpdatedAt; !got.Equal(want) {
			t.Fatalf("got %v, want %v", got, want)
		}

		// fetch user from database & compare
		if other, err := s.FindUserByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(uu, other) {
			t.Fatalf("mismatch: %#v != %#v", uu, other)
		}
	})

	t.Run("ErrUnauthorized", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, _ := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane"})
		_, ctx1 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Bob"})

		newName := "Jane Doe"

		if _, err := s.UpdateUser(ctx1, user0.ID, service.UserUpdate{Name: &newName}); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EUNAUTHORIZED {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EUNAUTHORIZED)
		}
	})

	t.Run("ErrNotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, ctx0 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane"})

		newName := "Jane Doe"

		if _, err := s.UpdateUser(ctx0, user0.ID+1, service.UserUpdate{Name: &newName}); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.ENOTFOUND {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.ENOTFOUND)
		}
	})

	t.Run("ErrCtxDone", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, ctx0 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane"})

		newName := "Jane Doe"

		ctxToCancel, cancel := context.WithCancel(ctx0)

		cancel()

		if _, err := s.UpdateUser(ctxToCancel, user0.ID, service.UserUpdate{Name: &newName}); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINTERNAL {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EINTERNAL)
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, ctx0 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane"})

		if err := s.DeleteUser(ctx0, user0.ID); err != nil {
			t.Fatal(err)
		} else if _, err := s.FindUserByID(context.Background(), 1); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.ENOTFOUND {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.ENOTFOUND)
		}
	})

	t.Run("ErrUnauthorized", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, _ := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane"})
		_, ctx1 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Bob"})

		if err := s.DeleteUser(ctx1, user0.ID); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EUNAUTHORIZED {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EUNAUTHORIZED)
		}
	})

	t.Run("ErrNotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, ctx0 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane"})

		if err := s.DeleteUser(ctx0, user0.ID+1); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.ENOTFOUND {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.ENOTFOUND)
		}
	})

	t.Run("ErrCtxDone", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, ctx0 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane"})

		ctxToCancel, cancel := context.WithCancel(ctx0)

		cancel()

		if err := s.DeleteUser(ctxToCancel, user0.ID); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINTERNAL {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EINTERNAL)
		}
	})
}

func TestUserService_FinUserByID(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, ctx0 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane", Email: null.StringFrom("jane@test.com")})

		if uu, err := s.FindUserByID(ctx0, user0.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(uu, user0) {
			t.Fatalf("mismatch: %#v != %#v", uu, user0)
		}

		if uu, err := s.FindUserByEmail(ctx0, user0.Email.String); err != nil {
			t.Fatal(err)
		} else if got, want := uu.ID, user0.ID; got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
	})

	t.Run("ErrNotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		if _, err := s.FindUserByID(context.Background(), 1); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.ENOTFOUND {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.ENOTFOUND)
		}
	})

	t.Run("ErrCtxDone", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		user0, ctx0 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane", Email: null.StringFrom("jane.doe@test.com")})

		ctxToCancel, cancel := context.WithCancel(ctx0)

		cancel()

		if _, err := s.FindUserByID(ctxToCancel, user0.ID); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINTERNAL {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EINTERNAL)
		}
	})
}

func TestUserService_FindUsers(t *testing.T) {

	t.Run("Name", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		ctx := context.Background()

		MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane", Email: null.StringFrom("jane@test.com")})
		MustCreateUser(t, context.Background(), db, &entity.User{Name: "Bob", Email: null.StringFrom("bob@test.com")})

		filterName := "Jane"

		if users, n, err := s.FindUsers(ctx, service.UserFilter{Name: &filterName}); err != nil {
			t.Fatal(err)
		} else if len(users) != 1 {
			t.Fatalf("got %d, want %d", len(users), 1)
		} else if got, want := users[0].Name, "Jane"; got != want {
			t.Fatalf("got %q, want %q", got, want)
		} else if got, want := n, 1; got != want {
			t.Fatalf("got %d, want %d", got, want)
		}

		filterName = "Bob"

		if users, n, err := s.FindUsers(ctx, service.UserFilter{Name: &filterName}); err != nil {
			t.Fatal(err)
		} else if len(users) != 1 {
			t.Fatalf("got %d, want %d", len(users), 1)
		} else if got, want := users[0].Name, "Bob"; got != want {
			t.Fatalf("got %q, want %q", got, want)
		} else if got, want := n, 1; got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
	})

	t.Run("Email", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		ctx := context.Background()

		MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane", Email: null.StringFrom("jane@test.com")})
		MustCreateUser(t, context.Background(), db, &entity.User{Name: "Bob", Email: null.StringFrom("bob@test.com")})

		filterEmail := "jane@test.com"

		if users, n, err := s.FindUsers(ctx, service.UserFilter{Email: &filterEmail}); err != nil {
			t.Fatal(err)
		} else if len(users) != 1 {
			t.Fatalf("got %d, want %d", len(users), 1)
		} else if got, want := users[0].Email.String, "jane@test.com"; got != want {
			t.Fatalf("got %q, want %q", got, want)
		} else if got, want := n, 1; got != want {
			t.Fatalf("got %d, want %d", got, want)
		}

		filterEmail = "bob@test.com"

		if users, n, err := s.FindUsers(ctx, service.UserFilter{Email: &filterEmail}); err != nil {
			t.Fatal(err)
		} else if len(users) != 1 {
			t.Fatalf("got %d, want %d", len(users), 1)
		} else if got, want := users[0].Email.String, "bob@test.com"; got != want {
			t.Fatalf("got %q, want %q", got, want)
		} else if got, want := n, 1; got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
	})

	t.Run("ErrCtxDone", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		// truncate users test table
		MustTruncateTable(t, db, "users")

		s := postgres.NewUserService(db)

		ctx := context.Background()

		MustCreateUser(t, context.Background(), db, &entity.User{Name: "Jane", Email: null.StringFrom("jane.doe@test.com")})

		ctxToCancel, cancel := context.WithCancel(ctx)

		cancel()

		if _, _, err := s.FindUsers(ctxToCancel, service.UserFilter{}); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINTERNAL {
			t.Fatalf("got %v, want %v", apperr.ErrorCode(err), apperr.EINTERNAL)
		}
	})
}

func MustCreateUser(tb testing.TB, ctx context.Context, db *postgres.DB, user *entity.User) (*entity.User, context.Context) {
	tb.Helper()
	if err := postgres.NewUserService(db).CreateUser(ctx, user); err != nil {
		tb.Fatal(err)
	}
	return user, app.NewContextWithUser(ctx, user)
}

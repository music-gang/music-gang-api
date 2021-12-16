package postgres_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/app/util"
	"github.com/music-gang/music-gang-api/postgres"
	"gopkg.in/guregu/null.v4"
)

func TestAuthService_CreateAuth(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth := &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.done@test.com"),
			},
		}

		if err := s.CreateAuth(context.Background(), auth); err != nil {
			t.Fatal(err)
		} else if got, want := auth.ID, int64(1); got != want {
			t.Fatalf("got id %d, want %d", got, want)
		} else if auth.CreatedAt.IsZero() {
			t.Fatal("got zero CreatedAt")
		} else if auth.UpdatedAt.IsZero() {
			t.Fatal("got zero UpdatedAt")
		}

		if other, err := s.FindAuthByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(auth, other) {
			t.Fatalf("mismatch: %#v != %#v", other, auth)
		}

		if user, err := postgres.NewUserService(db).FindUserByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if len(user.Auths) != 1 {
			t.Fatalf("got %d auths, want 1", len(user.Auths))
		} else if auth := user.Auths[0]; auth.ID != 1 || auth.UserID != user.ID {
			t.Fatalf("got auth %#v, want id 1 and user id %d", auth, user.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		auth0, ctx0 := MustCreateAuth(t, context.Background(), db, &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.doe@test.com"),
			},
		})

		auth01 := &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN-NEW"),
			RefreshToken: null.StringFrom("REFRESHTOKEN-NEW"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User:         auth0.User,
		}

		s := postgres.NewAuthService(db)

		if err := s.CreateAuth(ctx0, auth01); err != nil {
			t.Fatal(err)
		} else if got, want := auth01.ID, int64(1); got != want {
			t.Fatalf("got id %d, want %d", got, want)
		} else if auth0.UserID != auth01.UserID {
			t.Fatalf("got user id %d, want %d", auth0.UserID, auth01.UserID)
		}
	})

	t.Run("ErrSourceRequired", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth := &entity.Auth{
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.doe@test.com"),
			},
		}

		if err := s.CreateAuth(context.Background(), auth); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINVALID {
			t.Fatalf("got %v, want EINVALID", err)
		}
	})

	t.Run("ErrSourceIDRequired", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth := &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.doe@test.com"),
			},
		}

		if err := s.CreateAuth(context.Background(), auth); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINVALID {
			t.Fatalf("got %v, want EINVALID", err)
		}
	})

	t.Run("ErrUserRequired", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth := &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
		}

		if err := s.CreateAuth(context.Background(), auth); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINVALID {
			t.Fatalf("got %v, want EINVALID", err)
		}
	})

	t.Run("ErrAccessTokenEmpty", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth := &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom(""),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.doe@test.com"),
			},
		}

		if err := s.CreateAuth(context.Background(), auth); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINVALID {
			t.Fatalf("got %v, want EINVALID", err)
		}
	})

	t.Run("ErrRefreshTokenEmpty", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth := &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom(""),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.doe@test.com"),
			},
		}

		if err := s.CreateAuth(context.Background(), auth); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EINVALID {
			t.Fatalf("got %v, want EINVALID", err)
		}
	})

	t.Run("CannotCreateWithAlreadyUsedEmail", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth := &entity.Auth{
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				Name:     "Jane Doe",
				Email:    null.StringFrom("jane.doe@test.com"),
				Password: null.StringFrom("123456"),
			},
		}

		if err := s.CreateAuth(context.Background(), auth); err != nil {
			t.Fatal(err)
		}

		tamperAuth := &entity.Auth{
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				Name:     "Bob Smith",
				Email:    null.StringFrom("jane.doe@test.com"),
				Password: null.StringFrom("123456"),
			},
		}

		if err := s.CreateAuth(context.Background(), tamperAuth); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EFORBIDDEN {
			t.Fatalf("got %v, want EFORBIDDEN", err)
		}
	})

	t.Run("CannotCreateLocalAfterOauth", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth := &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.doe@test.com"),
			},
		}

		if err := s.CreateAuth(context.Background(), auth); err != nil {
			t.Fatal(err)
		}

		localAuth := &entity.Auth{
			Source: entity.AuthSourceGitHub,
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.doe@test.com"),
			},
		}

		if err := s.CreateAuth(context.Background(), localAuth); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EFORBIDDEN {
			t.Fatalf("got %v, want EFORBIDDEN", err)
		}
	})

	t.Run("CanCreateOauthAfterLocal", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		localAuth := &entity.Auth{
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				Name:     "Jane Doe",
				Email:    null.StringFrom("jane.doe@test.com"),
				Password: null.StringFrom("123456"),
			},
		}

		if err := s.CreateAuth(context.Background(), localAuth); err != nil {
			t.Fatal(err)
		}

		auth := &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.doe@test.com"),
			},
		}

		if err := s.CreateAuth(context.Background(), auth); err != nil {
			t.Fatal(err)
		}

		userIdFind := int64(1)

		if _, n, err := s.FindAuths(context.Background(), service.AuthFilter{
			UserID: &userIdFind,
		}); err != nil {
			t.Fatal(err)
		} else if n != 2 {
			t.Fatalf("got %d, want 2", n)
		}
	})
}

func TestAuthService_DeleteAuth(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth, ctx := MustCreateAuth(t, context.Background(), db, &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name: "Jane Doe",
			},
		})

		if err := s.DeleteAuth(ctx, auth.ID); err != nil {
			t.Fatal(err)
		} else if _, err := s.FindAuthByID(ctx, auth.ID); apperr.ErrorCode(err) != apperr.ENOTFOUND {
			t.Fatalf("got %v, want ENOTFOUND", err)
		}
	})

	t.Run("ErrNotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		if err := s.DeleteAuth(context.Background(), 1); apperr.ErrorCode(err) != apperr.ENOTFOUND {
			t.Fatalf("got %v, want ENOTFOUND", err)
		}
	})

	t.Run("ErrForbidden", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		user, _ := MustCreateUser(t, context.Background(), db, &entity.User{
			Name: "Jane Doe",
		})

		auth, ctx := MustCreateAuth(t, context.Background(), db, &entity.Auth{
			Source:       entity.AuthSourceLocal,
			SourceID:     null.StringFromPtr(nil),
			AccessToken:  null.StringFromPtr(nil),
			RefreshToken: null.StringFromPtr(nil),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User:         user,
		})

		if err := s.DeleteAuth(ctx, auth.ID); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EFORBIDDEN {
			t.Fatalf("got %v, want EFORBIDDEN", err)
		}
	})

	t.Run("ErrUnauthorized", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth0, _ := MustCreateAuth(t, context.Background(), db, &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name: "Jane Doe",
			},
		})

		_, ctx1 := MustCreateAuth(t, context.Background(), db, &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID-1"),
			AccessToken:  null.StringFrom("ACCESSTOKEN-1"),
			RefreshToken: null.StringFrom("REFRESHTOKEN-1"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name: "Bob Doe",
			},
		})

		if err := s.DeleteAuth(ctx1, auth0.ID); err == nil {
			t.Fatal("expected error")
		} else if apperr.ErrorCode(err) != apperr.EUNAUTHORIZED {
			t.Fatalf("got %v, want EUNAUTHORIZED", err)
		}
	})
}

func TestAuthService_FindAuthByID(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		auth, ctx := MustCreateAuth(t, context.Background(), db, &entity.Auth{
			Source:       entity.AuthSourceGitHub,
			SourceID:     null.StringFrom("SOURCEID"),
			AccessToken:  null.StringFrom("ACCESSTOKEN"),
			RefreshToken: null.StringFrom("REFRESHTOKEN"),
			Expiry:       null.TimeFrom(util.AppNowUTC()),
			User: &entity.User{
				Name: "Jane Doe",
			},
		})

		if _, err := s.FindAuthByID(ctx, auth.ID); err != nil {
			t.Fatal(err)
		}

		if other, err := s.FindAuthByID(ctx, auth.ID+1); apperr.ErrorCode(err) != apperr.ENOTFOUND {
			t.Fatalf("%#v", other)
			t.Fatalf("got %v, want ENOTFOUND", err)
		}
	})
}

func TestAuthService_FindAuths(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		TruncateTablesForAuthTests(t, db)

		s := postgres.NewAuthService(db)

		ctx := context.Background()

		MustCreateAuth(t, context.Background(), db, &entity.Auth{
			Source:      "SRCA",
			SourceID:    null.StringFrom("X1"),
			AccessToken: null.StringFrom("ACCESSX1"),
			User:        &entity.User{Name: "X", Email: null.StringFrom("x@y.com")},
		})
		MustCreateAuth(t, context.Background(), db, &entity.Auth{
			Source:      "SRCB",
			SourceID:    null.StringFrom("X2"),
			AccessToken: null.StringFrom("ACCESSX2"),
			User:        &entity.User{Name: "X", Email: null.StringFrom("x@y.com")},
		})
		MustCreateAuth(t, context.Background(), db, &entity.Auth{
			Source:      entity.AuthSourceGitHub,
			SourceID:    null.StringFrom("Y"),
			AccessToken: null.StringFrom("ACCESSY"),
			User:        &entity.User{Name: "Y"},
		})

		userID := int64(1)
		if a, n, err := s.FindAuths(ctx, service.AuthFilter{UserID: &userID}); err != nil {
			t.Fatal(err)
		} else if got, want := len(a), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if got, want := a[0].SourceID, "X1"; got.String != want {
			t.Fatalf("[]=%v, want %v", got, want)
		} else if got, want := a[1].SourceID, "X2"; got.String != want {
			t.Fatalf("[]=%v, want %v", got, want)
		} else if got, want := n, 2; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})
}

func MustCreateAuth(tb testing.TB, ctx context.Context, db *postgres.DB, auth *entity.Auth) (*entity.Auth, context.Context) {
	tb.Helper()

	s := postgres.NewAuthService(db)

	if err := s.CreateAuth(ctx, auth); err != nil {
		tb.Fatal(err)
	}

	return auth, app.NewContextWithUser(ctx, auth.User)
}

func TruncateTablesForAuthTests(tb testing.TB, db *postgres.DB) {
	tb.Helper()

	MustTruncateTable(tb, db, "users")
	MustTruncateTable(tb, db, "auths")
}

package http_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/mock"
)

func TestMiddleware_JWT(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {

				if token != "OK" {
					return nil, apperr.Errorf(apperr.EINVALID, "invalid token")
				}

				return &entity.AppClaims{
					Auth: &entity.Auth{
						ID:   1,
						User: &entity.User{ID: 1},
					},
				}, nil
			},
		}

		s.UserService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				return &entity.User{ID: 1}, nil
			},
		}

		s.AuthService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				return &entity.Auth{ID: 1}, nil
			},
		}

		req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/user", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+"OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("ErrParseToken", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "invalid token")
			},
		}

		req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/user", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("ErrUserNotFound", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				return &entity.AppClaims{
					Auth: &entity.Auth{
						ID:   1,
						User: &entity.User{ID: 1},
					},
				}, nil
			},
		}

		s.UserService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/user", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("ErrAuthNotFound", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				return &entity.AppClaims{
					Auth: &entity.Auth{
						ID:   1,
						User: &entity.User{ID: 1},
					},
				}, nil
			},
		}

		s.UserService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				return &entity.User{ID: 1}, nil
			},
		}

		s.AuthService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/user", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

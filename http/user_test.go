package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/mock"
)

func TestUserHandler(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/user", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var userData map[string]interface{}

		if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
			t.Fatal(err)
		}

		if u, ok := userData["user"]; !ok {
			t.Error("expected user key in userData")
		} else if user, ok := u.(map[string]interface{}); !ok {
			t.Error("expected user to be a map")
		} else if id, ok := user["id"]; !ok {
			t.Error("expected id key in user")
		} else if id != float64(1) {
			t.Errorf("expected id %f, got %f", float64(1), id)
		}
	})
}

func TestUser_Update(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}
				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}
				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}
		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}
				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}
		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			UserService: &mock.UserService{
				UpdateUserFn: func(ctx context.Context, id int64, user service.UserUpdate) (*entity.User, error) {
					if id == 1 {
						updatedUser := &entity.User{ID: 1}
						if user.Name != nil {
							updatedUser.Name = *user.Name
						}
						return updatedUser, nil
					}
					return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
				},
			},
		}

		req, err := http.NewRequest(http.MethodPut, s.URL()+"/v1/user", strings.NewReader(`{"name":"updated"}`))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var userData map[string]interface{}

		if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
			t.Fatal(err)
		}

		if u, ok := userData["user"]; !ok {
			t.Error("expected user key in userData")
		} else if user, ok := u.(map[string]interface{}); !ok {
			t.Error("expected user to be a map")
		} else if id, ok := user["id"]; !ok {
			t.Error("expected id key in user")
		} else if id != float64(1) {
			t.Errorf("expected id %f, got %f", float64(1), id)
		} else if name, ok := user["name"]; !ok {
			t.Error("expected name key in user")
		} else if name != "updated" {
			t.Errorf("expected name %s, got %s", "updated", name)
		}
	})

	t.Run("UpdateErr", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}
				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}
				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}
		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			UserService: &mock.UserService{
				UpdateUserFn: func(ctx context.Context, id int64, user service.UserUpdate) (*entity.User, error) {
					return nil, apperr.Errorf(apperr.EINTERNAL, "internal error")
				},
			},
		}

		req, err := http.NewRequest(http.MethodPut, s.URL()+"/v1/user", strings.NewReader(`{"name":"updated"}`))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

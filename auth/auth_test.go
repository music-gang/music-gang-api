package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/util"
	"github.com/music-gang/music-gang-api/auth"
	"github.com/music-gang/music-gang-api/config"
	"github.com/music-gang/music-gang-api/mock"
	"gopkg.in/guregu/null.v4"
)

var authSourceLocal = entity.AuthSourceLocal
var authSourceGithub = entity.AuthSourceGitHub

type Auth struct {
	*auth.AuthService

	as mock.AuthService
	us mock.UserService
}

func NewAuth() *Auth {
	a := &Auth{}

	config.LoadConfigWithOptions(config.LoadOptions{ConfigFilePath: "../config.yaml"})

	a.AuthService = auth.NewAuth(&a.as, &a.us, config.GetConfig().TEST.Auths)

	return a
}

func setupMockOAuthServer(t testing.TB) (*httptest.Server, func()) {

	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		// Should return authorization code back to the user
	})

	mux.HandleFunc("/github/token", func(w http.ResponseWriter, r *http.Request) {
		// Should return acccess token back to the user
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": "fake_access_token"}`))
	})

	server := httptest.NewServer(mux)

	return server, func() {
		server.Close()
	}
}

func TestAuth_TestLocalProvider(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := NewAuth()

		name := "Jane Doe"
		email := "jane.done@test.com"
		password := "123456"

		var hashedPassword []byte
		if p, err := util.HashPassword(password); err == nil {
			hashedPassword = p
		}

		s.as.CreateAuthFn = func(ctx context.Context, auth *entity.Auth) error {
			return nil
		}

		s.us.FindUserByEmailFn = func(ctx context.Context, email string) (*entity.User, error) {
			return &entity.User{
				ID:        1,
				Name:      name,
				Email:     null.StringFrom(email),
				Password:  null.StringFrom(string(hashedPassword)),
				CreatedAt: util.AppNowUTC(),
				UpdatedAt: util.AppNowUTC(),
				Auths: []*entity.Auth{{
					ID:        1,
					UserID:    1,
					Source:    authSourceLocal,
					CreatedAt: util.AppNowUTC(),
					UpdatedAt: util.AppNowUTC(),
				}},
			}, nil
		}

		if auth, err := s.Auhenticate(context.Background(), &entity.AuthUserOptions{
			Source: &authSourceLocal,
			UserParams: &entity.UserParams{
				Email:    &email,
				Password: &password,
			},
		}); err != nil {
			t.Errorf("Expected no error, got %v", err)
		} else if auth.User.Name != "Jane Doe" {
			t.Errorf("Expected user name to be 'Jane Doe', got %v", auth.User.Name)
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {

		s := NewAuth()

		email := "jane.done@test.com"
		password := "123456"

		s.us.FindUserByEmailFn = func(ctx context.Context, email string) (*entity.User, error) {
			return nil, apperr.Errorf(apperr.ENOTFOUND, "User not found")
		}

		if _, err := s.Auhenticate(context.Background(), &entity.AuthUserOptions{
			Source: &authSourceLocal,
			UserParams: &entity.UserParams{
				Email:    &email,
				Password: &password,
			},
		}); err == nil {
			t.Errorf("Expected error, got nil")
		} else if apperr.ErrorCode(err) != apperr.ENOTFOUND {
			t.Errorf("Expected error code to be %v, got %v", apperr.ENOTFOUND, err)
		}
	})

	t.Run("PasswordMismatch", func(t *testing.T) {

		s := NewAuth()

		name := "Jane Doe"
		email := "jane.done@test.com"
		password := "123456"

		var hashedPassword []byte
		if p, err := util.HashPassword(password); err == nil {
			hashedPassword = p
		}

		s.as.CreateAuthFn = func(ctx context.Context, auth *entity.Auth) error {
			return nil
		}

		s.us.FindUserByEmailFn = func(ctx context.Context, email string) (*entity.User, error) {
			return &entity.User{
				ID:        1,
				Name:      name,
				Email:     null.StringFrom(email),
				Password:  null.StringFrom(string(hashedPassword)),
				CreatedAt: util.AppNowUTC(),
				UpdatedAt: util.AppNowUTC(),
				Auths: []*entity.Auth{{
					ID:        1,
					UserID:    1,
					Source:    authSourceLocal,
					CreatedAt: util.AppNowUTC(),
					UpdatedAt: util.AppNowUTC(),
				}},
			}, nil
		}

		wrongPassword := "wrong_password"

		if _, err := s.Auhenticate(context.Background(), &entity.AuthUserOptions{
			Source: &authSourceLocal,
			UserParams: &entity.UserParams{
				Email:    &email,
				Password: &wrongPassword,
			}}); err == nil {
			t.Errorf("Expected error, got nil")
		} else if apperr.ErrorCode(err) != apperr.EUNAUTHORIZED {
			t.Errorf("Expected error code to be %v, got %v", apperr.EUNAUTHORIZED, err)
		}
	})
}

func TestAuth_TestGithubProvider(t *testing.T) {

	server, close := setupMockOAuthServer(t)
	defer close()

	t.Run("OK", func(t *testing.T) {

		s := NewAuth()

		githubProvider := s.ProviderBySource(authSourceGithub).(*auth.GithubProvider)
		githubProvider.MockUserFn()

		githubProvider.SetConfig(config.AuthConfig{
			ClientID:     "CLIENT_ID",
			ClientSecret: "CLIENT_SECRET",
			RedirectURL:  "REDIRECT_URL",
			Scopes:       []string{},
			Endpoint: struct {
				AuthURL   string "yaml:\"auth_url\""
				TokenURL  string "yaml:\"token_url\""
				AuthStyle int    "yaml:\"auth_style\""
			}{
				AuthURL:  server.URL + "/auth",
				TokenURL: server.URL + "/github/token",
			},
		})

		s.as.CreateAuthFn = func(ctx context.Context, auth *entity.Auth) error {
			return nil
		}

		fakeAuthCode := "FAKE_AUTH_CODE"

		if auth, err := s.Auhenticate(context.Background(), &entity.AuthUserOptions{
			Source:   &authSourceGithub,
			AuthCode: &fakeAuthCode,
		}); err != nil {
			t.Errorf("Expected no error, got %v", err)
		} else if auth.User.Name != "Jane Doe" {
			t.Errorf("Expected user name to be 'Jane Doe', got %v", auth.User.Name)
		}
	})
}

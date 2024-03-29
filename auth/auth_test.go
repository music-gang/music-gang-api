package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/auth"
	"github.com/music-gang/music-gang-api/common"
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

	a.AuthService = auth.NewAuth(&a.as, &a.us, config.GetConfig().APP.Auths)

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

		name := "JaneDoe"
		email := "jane.done@test.com"
		password := "123456"

		var hashedPassword []byte
		if p, err := common.HashPassword(password); err == nil {
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
				CreatedAt: common.AppNowUTC(),
				UpdatedAt: common.AppNowUTC(),
				Auths: []*entity.Auth{{
					ID:        1,
					UserID:    1,
					Source:    authSourceLocal,
					CreatedAt: common.AppNowUTC(),
					UpdatedAt: common.AppNowUTC(),
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
		} else if auth.User.Name != "JaneDoe" {
			t.Errorf("Expected user name to be 'JaneDoe', got %v", auth.User.Name)
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

		name := "JaneDoe"
		email := "jane.done@test.com"
		password := "123456"

		var hashedPassword []byte
		if p, err := common.HashPassword(password); err == nil {
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
				CreatedAt: common.AppNowUTC(),
				UpdatedAt: common.AppNowUTC(),
				Auths: []*entity.Auth{{
					ID:        1,
					UserID:    1,
					Source:    authSourceLocal,
					CreatedAt: common.AppNowUTC(),
					UpdatedAt: common.AppNowUTC(),
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
				AuthURL   string `env:"AUTH_URL"`
				TokenURL  string `env:"TOKEN_URL"`
				AuthStyle int    `env:"AUTH_STYLE"`
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
		} else if auth.User.Name != "JaneDoe" {
			t.Errorf("Expected user name to be 'JaneDoe', got %v", auth.User.Name)
		}
	})
}

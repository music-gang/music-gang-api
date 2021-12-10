package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/music-gang/music-gang-api/app/entity"
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
}

func NewAuth() *Auth {
	a := &Auth{}

	config.LoadConfigWithOptions(config.LoadOptions{ConfigFilePath: "../config.yaml"})

	a.AuthService = auth.NewAuth(&a.as, config.GetConfig().TEST.Auths)
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

		s.as.CreateAuthFn = func(ctx context.Context, auth *entity.Auth) error {
			return nil
		}

		if auth, err := s.Auhenticate(context.Background(), &entity.AuthUserOptions{
			Source: &authSourceLocal,
			User: &entity.User{
				Name:  "Jane Doe",
				Email: null.StringFrom("jane.doe@test.com"),
			},
		}); err != nil {
			t.Errorf("Expected no error, got %v", err)
		} else if auth.User.Name != "Jane Doe" {
			t.Errorf("Expected user name to be 'Jane Doe', got %v", auth.User.Name)
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

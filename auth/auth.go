package auth

import (
	"context"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"golang.org/x/oauth2"
)

// AuthUserOptions represents the options for a user during fetching auth service.
type AuthUserOptions struct {
	AuthCode *string
	Source   *string
	User     *entity.User
}

// AuthProvider is the interface for the auth provider.
type AuthProvider interface {
	// GetAuthURL returns the auth url, nil if not supported.
	GetOAuthConfig() *oauth2.Config
	// Source returns the source of the auth provider.
	Source() string
	// GetUser returns the user from the auth provider.
	// May returns err if something went wrong.
	User(ctx context.Context, opts *AuthUserOptions) (*entity.Auth, error)
}

type Auth struct {
	authService service.AuthService
	providers   map[string]AuthProvider
}

// NewAuth creates a new Auth instance
func NewAuth(authService service.AuthService) *Auth {
	auth := &Auth{authService: authService}
	auth.initProviders()
	return auth
}

// Auhenticate authenticates the user with the given auth source and options.
// Each provider has its own options and implements its validation.
func (a *Auth) Auhenticate(ctx context.Context, opts *AuthUserOptions) (*entity.Auth, error) {

	if opts == nil || opts.Source == nil || *opts.Source == "" {
		return nil, apperr.Errorf(apperr.EINVALID, "source is required")
	}

	if p := a.providers[*opts.Source]; p == nil {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "auth provider not found")
	}

	auth, err := a.providers[*opts.Source].User(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := a.authService.CreateAuth(ctx, auth); err != nil {
		return nil, err
	}

	return auth, nil
}

// initProviders initializes the supported providers.
func (a *Auth) initProviders() {
	a.providers = make(map[string]AuthProvider)

	a.providers[entity.AuthSourceGitHub] = NewGithubProvider(&oauth2.Config{})
	a.providers[entity.AuthSourceLocal] = NewLocalProvider()
}

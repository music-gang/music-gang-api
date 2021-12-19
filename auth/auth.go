package auth

import (
	"context"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/config"
	"golang.org/x/oauth2"
)

// AuthProvider is the interface for the auth provider.
type AuthProvider interface {
	// Auhenticate returns the user from the auth provider.
	// May returns err if something went wrong.
	Auhenticate(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error)
	// GetAuthURL returns the auth url, nil if not supported.
	GetOAuthConfig() *oauth2.Config
	// Source returns the source of the auth provider.
	Source() string
}

var _ service.AuthService = (*AuthService)(nil)

// Auth is the auth service.
// Contains the supported providers.
// Contains the underlying auth service. maybe a sql db.
// Implements service.AuthService interface.
type AuthService struct {
	as        service.AuthService
	us        service.UserService
	providers map[string]AuthProvider

	conf config.AuthListConfig
}

// NewAuth creates a new Auth instance
func NewAuth(authService service.AuthService, userService service.UserService, conf config.AuthListConfig) *AuthService {
	auth := &AuthService{as: authService, us: userService, conf: conf}
	auth.initProviders()
	return auth
}

// Auhenticate authenticates the user with the given auth source and options.
// Each provider has its own options and implements its validation.
func (a *AuthService) Auhenticate(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {

	if opts == nil || opts.Source == nil || *opts.Source == "" {
		return nil, apperr.Errorf(apperr.EINVALID, "source is required")
	}

	if p := a.providers[*opts.Source]; p == nil {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "auth provider not found")
	}

	auth, err := a.providers[*opts.Source].Auhenticate(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := a.CreateAuth(ctx, auth); err != nil {
		return nil, err
	}

	return auth, nil
}

// CreateAuth creates a new auth.
// If is attached to a user, links the auth to the user, otherwise creates a new user.
// On success, the auth.ID is set.
func (a *AuthService) CreateAuth(ctx context.Context, auth *entity.Auth) error {
	return a.as.CreateAuth(ctx, auth)
}

// DeleteAuth deletes an auth.
// Do not delete underlying user.
func (a *AuthService) DeleteAuth(ctx context.Context, id int64) error {
	return a.as.DeleteAuth(ctx, id)
}

// FindAuthByID returns a single auth by its id.
func (a *AuthService) FindAuthByID(ctx context.Context, id int64) (*entity.Auth, error) {
	return a.as.FindAuthByID(ctx, id)
}

// FindAuths returns a list of auths.
// Predicate can be used to filter the results.
// Also returns the total count of auths.
func (a *AuthService) FindAuths(ctx context.Context, filter service.AuthFilter) (entity.Auths, int, error) {
	return a.as.FindAuths(ctx, filter)
}

// initProviders initializes the supported providers.
func (a *AuthService) initProviders() {
	a.providers = make(map[string]AuthProvider)

	a.providers[entity.AuthSourceGitHub] = NewGithubProvider(a.conf.Github, a.as)
	a.providers[entity.AuthSourceLocal] = NewLocalProvider(a.as, a.us)
}

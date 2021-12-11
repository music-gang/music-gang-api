package auth

import (
	"context"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

var _ AuthProvider = (*LocalProvider)(nil)

// LocalProvider implements the AuthProvider interface for local authentication.
type LocalProvider struct {
	authService service.AuthService
	userService service.UserService
}

// NewLocalProvider returns a new LocalProvider.
func NewLocalProvider(authService service.AuthService, userService service.UserService) *LocalProvider {
	return &LocalProvider{authService: authService, userService: userService}
}

//  GetConfig returns the oauth2.Config for the provider.
// Local provider does not use oauth2.
func (p *LocalProvider) GetOAuthConfig() *oauth2.Config { return nil }

// Source returns the source of the provider.
func (p *LocalProvider) Source() string {
	return entity.AuthSourceLocal
}

// Auhenticate returns the user from local auth.
// Requires a user passed through opts.User.
func (p *LocalProvider) Auhenticate(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {

	if opts == nil || opts.UserParams == nil {
		return nil, apperr.Errorf(apperr.EINVALID, "opts.UserOptions is required")
	}

	if opts.UserParams.Email == nil || *opts.UserParams.Email == "" ||
		opts.UserParams.Password == nil || *opts.UserParams.Password == "" {
		return nil, apperr.Errorf(apperr.EINVALID, "email and password are required")
	}

	user, err := p.userService.FindUserByEmail(ctx, *opts.UserParams.Email)
	if err != nil {
		return nil, err
	}

	var auth *entity.Auth

	for _, userAuth := range user.Auths {
		if userAuth.Source == entity.AuthSourceLocal {
			auth = userAuth
			break
		}
	}

	if auth == nil {
		return nil, apperr.Errorf(apperr.EINVALID, "user does not have local auth")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(*opts.UserParams.Password)); err != nil {
		return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "wrong credentials")
	}

	auth.User = user

	return auth, nil
}

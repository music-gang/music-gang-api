package auth

import (
	"context"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"golang.org/x/oauth2"
	"gopkg.in/guregu/null.v4"
)

var _ AuthProvider = (*LocalProvider)(nil)

// LocalProvider implements the AuthProvider interface for local authentication.
type LocalProvider struct{}

// NewLocalProvider returns a new LocalProvider.
func NewLocalProvider() *LocalProvider {
	return &LocalProvider{}
}

//  GetConfig returns the oauth2.Config for the provider.
// Local provider does not use oauth2.
func (p *LocalProvider) GetOAuthConfig() *oauth2.Config { return nil }

// Source returns the source of the provider.
func (p *LocalProvider) Source() string {
	return entity.AuthSourceLocal
}

// User returns the user from local auth.
// Requires a user passed through opts.User.
func (p *LocalProvider) User(ctx context.Context, opts *AuthUserOptions) (*entity.Auth, error) {

	if opts == nil || opts.User == nil {
		return nil, apperr.Errorf(apperr.EINVALID, "opts.User is required")
	} else if err := opts.User.Validate(); err != nil {
		return nil, err
	}

	auth := &entity.Auth{
		User:     opts.User,
		Source:   entity.AuthSourceLocal,
		SourceID: null.StringFromPtr(nil),
	}

	return auth, nil
}

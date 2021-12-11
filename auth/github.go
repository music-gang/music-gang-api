package auth

import (
	"context"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/config"
	"golang.org/x/oauth2"
	"gopkg.in/guregu/null.v4"
)

var _ AuthProvider = (*GithubProvider)(nil)

// GithubProvider is the Github implementation of AuthProvider.
type GithubProvider struct {
	config      config.AuthConfig
	authService service.AuthService
	userFn      func(ctx context.Context, client *github.Client) (*github.User, *github.Response, error)
}

// NewGithubProvider returns a new GithubProvider.
func NewGithubProvider(config config.AuthConfig, authService service.AuthService) *GithubProvider {
	return &GithubProvider{config: config, authService: authService, userFn: user}
}

//  GetConfig returns the oauth2.Config for the provider.
func (p *GithubProvider) GetOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   p.config.Endpoint.AuthURL,
			TokenURL:  p.config.Endpoint.TokenURL,
			AuthStyle: oauth2.AuthStyle(p.config.Endpoint.AuthStyle),
		},
		RedirectURL: p.config.RedirectURL,
		Scopes:      p.config.Scopes,
	}
}

// Source returns the source of the provider.
func (p *GithubProvider) Source() string {
	return entity.AuthSourceGitHub
}

// Auhenticate implements oauth2 for github.
// AuthUserOptions.AuthCode is required to exchange for tokens.
func (p *GithubProvider) Auhenticate(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {

	if opts == nil || opts.AuthCode == nil || *opts.AuthCode == "" {
		return nil, apperr.Errorf(apperr.EINVALID, "opts.AuthCode is required")
	}

	tok, err := p.GetOAuthConfig().Exchange(ctx, *opts.AuthCode)
	if err != nil {
		return nil, apperr.Errorf(apperr.EINVALID, "failed to exchange auth code for token: %v", err)
	}

	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tok.AccessToken},
	)))

	u, _, err := p.userFn(ctx, client)
	if err != nil {
		return nil, err
	}

	var name string
	if u.Name != nil {
		name = *u.Name
	} else if u.Login != nil {
		name = *u.Login
	}

	var tempEmail string
	if u.Email != nil {
		tempEmail = *u.Email
	}

	var email null.String

	if tempEmail == "" {
		email = null.StringFromPtr(nil)
	} else {
		email = null.StringFrom(tempEmail)
	}

	auth := &entity.Auth{
		Source:       entity.AuthSourceGitHub,
		SourceID:     null.StringFrom(fmt.Sprint(*u.ID)),
		AccessToken:  null.StringFrom(tok.AccessToken),
		RefreshToken: null.StringFrom(tok.RefreshToken),
		User: &entity.User{
			Name:  name,
			Email: email,
		},
	}

	if !tok.Expiry.IsZero() {
		auth.Expiry = null.TimeFrom(tok.Expiry)
	}

	if err := p.authService.CreateAuth(ctx, auth); err != nil {
		return nil, err
	}

	return auth, nil
}

// user implements oauth2 for github.
// It can be mooked for testing.
func user(ctx context.Context, client *github.Client) (*github.User, *github.Response, error) {

	u, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, nil, apperr.Errorf(apperr.EINTERNAL, "failed to get user: %v", err)
	} else if u.ID == nil || *u.ID == 0 {
		return nil, nil, apperr.Errorf(apperr.EINTERNAL, "User ID returned from Github is invalid, cannot authenticate")
	}

	return u, resp, nil
}

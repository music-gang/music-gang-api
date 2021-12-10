package auth

import (
	"context"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

func (a *Auth) ProviderBySource(source string) AuthProvider {
	return a.providers[source]
}

func (p *GithubProvider) SetConfig(config *oauth2.Config) {
	p.config = config
}

func (p *GithubProvider) MockUserFn() {
	p.userFn = func(ctx context.Context, client *github.Client) (*github.User, *github.Response, error) {
		return &github.User{
			ID:    github.Int64(1),
			Login: github.String("Jane Doe"),
			Name:  github.String("Jane Doe"),
			Email: github.String("jane.doe@test.com"),
		}, &github.Response{}, nil
	}
}

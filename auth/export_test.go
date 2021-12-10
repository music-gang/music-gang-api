package auth

import (
	"context"

	"github.com/google/go-github/v32/github"
	"github.com/music-gang/music-gang-api/config"
)

func (a *AuthService) ProviderBySource(source string) AuthProvider {
	return a.providers[source]
}

func (p *GithubProvider) SetConfig(config config.AuthConfig) {
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

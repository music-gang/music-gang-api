package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	apphttp "github.com/music-gang/music-gang-api/http"
	"github.com/music-gang/music-gang-api/mock"
)

var validPassword = "SecurePassword@123!"

type RegisterCase struct {
	Name   string
	Params apphttp.RegisterParams
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func TestAuth_Register(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		var authenticatedAuth *entity.Auth

		s.AuthService = &mock.AuthService{
			AuthentcateFn: func(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {
				return authenticatedAuth, nil
			},
			CreateAuthFn: func(ctx context.Context, auth *entity.Auth) error {
				auth.ID = 1
				auth.User.ID = 1
				auth.UserID = 1
				authenticatedAuth = auth
				return nil
			},
		}
		s.JWTService = &mock.JWTService{
			ExchangeFn: func(ctx context.Context, auth *entity.Auth) (*entity.TokenPair, error) {
				return &entity.TokenPair{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
					TokenType:    "Bearer",
					Expiry:       3600,
				}, nil
			},
		}

		registerParam := apphttp.RegisterParams{
			Email:           "jane.doe@test.com",
			Name:            "Jane Doe",
			Password:        validPassword,
			ConfirmPassword: validPassword,
		}

		jsonValue := MustMarshalJSON(t, registerParam)

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/auth/register", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var reponseData LoginResponse

		if err := json.NewDecoder(resp.Body).Decode(&reponseData); err != nil {
			t.Fatal(err)
		}

		if reponseData.AccessToken == "" {
			t.Error("expected access token, got empty string")
		}

		if reponseData.RefreshToken == "" {
			t.Error("expected refresh token, got empty string")
		}

		if reponseData.ExpiresIn <= 0 {
			t.Errorf("expected expires_in > 0, got %d", reponseData.ExpiresIn)
		}

		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}
	})

	t.Run("ErrMissingEmail", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		registerParam := apphttp.RegisterParams{
			Name:            "Jane Doe",
			Password:        validPassword,
			ConfirmPassword: validPassword,
		}

		jsonValue := MustMarshalJSON(t, registerParam)

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/auth/register", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrInvalidEmail", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		registerParam := apphttp.RegisterParams{
			Email:           "jane.doe.com",
			Name:            "Jane Doe",
			Password:        validPassword,
			ConfirmPassword: validPassword,
		}

		jsonValue := MustMarshalJSON(t, registerParam)

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/auth/register", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrMissingName", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		registerParam := apphttp.RegisterParams{
			Email:           "jane.doe@test.com",
			Password:        validPassword,
			ConfirmPassword: validPassword,
		}

		jsonValue := MustMarshalJSON(t, registerParam)

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/auth/register", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrMissingPassword", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		registerParam := apphttp.RegisterParams{
			Email: "jane.doe@test.com",
			Name:  "Jane Doe",
		}

		jsonValue := MustMarshalJSON(t, registerParam)

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/auth/register", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrNotValidPassword", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		registerParam := apphttp.RegisterParams{
			Email:    "jane.doe@test.com",
			Name:     "Jane Doe",
			Password: "not-secure-password",
		}

		jsonValue := MustMarshalJSON(t, registerParam)

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/auth/register", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrNotMatchingPasswords", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		registerParam := apphttp.RegisterParams{
			Email:           "jane.doe@test.com",
			Name:            "Jane Doe",
			Password:        validPassword,
			ConfirmPassword: "not-matching-password",
		}

		jsonValue := MustMarshalJSON(t, registerParam)

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/auth/register", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrCreateAuth", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.AuthService = &mock.AuthService{
			CreateAuthFn: func(ctx context.Context, auth *entity.Auth) error {
				return apperr.Errorf(apperr.EUNAUTHORIZED, "authentication failed")
			},
		}

		registerParam := apphttp.RegisterParams{
			Email:           "jane.doe@test.com",
			Name:            "Jane Doe",
			Password:        validPassword,
			ConfirmPassword: validPassword,
		}

		jsonValue := MustMarshalJSON(t, registerParam)

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/auth/register", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

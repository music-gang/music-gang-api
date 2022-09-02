package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/common"
	"github.com/music-gang/music-gang-api/mock"
)

var testFuel uint64 = uint64(entity.FuelMidActionAmount)

var contractRequestBody = `
{
	"name": "test contract",
	"description": "test contract",
	"user_id": 1,
	"visibility": "public",
	"max_fuel": ` + fmt.Sprint(testFuel) + `
}
`

var revisionRequestBody = `
{
	"version": "` + fmt.Sprint(entity.AnchorageVersion) + `",
	"notes": "test revision",
	"max_fuel": ` + fmt.Sprint(testFuel) + `
}
`

type revisionResponse struct {
	Revision *entity.Revision `json:"revision"`
}

func TestContract_ContractCreateHandler(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ContractService: &mock.ContractService{
				CreateContractFn: func(ctx context.Context, contract *entity.Contract) error {
					contract.ID = 1
					if err := contract.Validate(); err != nil {
						return err
					}
					return nil
				},
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract", bytes.NewBufferString(contractRequestBody))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		contractData := make(map[string]any)

		if err := json.NewDecoder(resp.Body).Decode(&contractData); err != nil {
			t.Fatal(err)
		} else if contractData["contract"] == nil {
			t.Fatalf("expected contract data, got %v", contractData)
		}
	})

	t.Run("InvalidRequest", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract", bytes.NewBufferString(contractRequestBody))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrCreateContract", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ContractService: &mock.ContractService{
				CreateContractFn: func(ctx context.Context, contract *entity.Contract) error {
					return apperr.Errorf(apperr.EINTERNAL, "internal error")
				},
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract", bytes.NewBufferString(contractRequestBody))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestContract_ContractUpdateHandler(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ContractService: &mock.ContractService{
				UpdateContractFn: func(ctx context.Context, id int64, contract service.ContractUpdate) (*entity.Contract, error) {
					return &entity.Contract{
						ID:          1,
						Name:        "test contract updated",
						Description: "test contract updated",
						UserID:      1,
						Visibility:  "public",
						MaxFuel:     entity.Fuel(testFuel),
					}, nil
				},
			},
		}

		req, err := http.NewRequest(http.MethodPut, s.URL()+"/v1/contract/1", bytes.NewBufferString(contractRequestBody))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		contractData := make(map[string]any)

		if err := json.NewDecoder(resp.Body).Decode(&contractData); err != nil {
			t.Fatal(err)
		} else if contractData["contract"] == nil {
			t.Fatalf("expected contract data, got %v", contractData)
		}
	})

	t.Run("InvalidContractID", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		req, err := http.NewRequest(http.MethodPut, s.URL()+"/v1/contract/invalid_id", bytes.NewBufferString(contractRequestBody))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("InvalidRequest", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		req, err := http.NewRequest(http.MethodPut, s.URL()+"/v1/contract/1", bytes.NewBufferString(contractRequestBody))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrContractUpdate", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ContractService: &mock.ContractService{
				UpdateContractFn: func(ctx context.Context, id int64, contract service.ContractUpdate) (*entity.Contract, error) {
					return nil, apperr.Errorf(apperr.EINTERNAL, "internal error")
				},
			},
		}

		req, err := http.NewRequest(http.MethodPut, s.URL()+"/v1/contract/1", bytes.NewBufferString(contractRequestBody))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestContract_ContractHandler(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindContractByIDFn: func(ctx context.Context, id int64) (*entity.Contract, error) {
				return &entity.Contract{
					ID:          1,
					Name:        "test contract",
					Description: "test contract",
					UserID:      1,
					Visibility:  "public",
					MaxFuel:     entity.Fuel(testFuel),
				}, nil
			},
		}

		req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/contract/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		contractData := make(map[string]any)

		if err := json.NewDecoder(resp.Body).Decode(&contractData); err != nil {
			t.Fatal(err)
		} else if contractData["contract"] == nil {
			t.Fatalf("expected contract data, got %v", contractData)
		}
	})

	t.Run("InvalidContractID", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/contract/invalid_id", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrFindContractByID", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindContractByIDFn: func(ctx context.Context, id int64) (*entity.Contract, error) {
				return nil, apperr.Errorf(apperr.EINTERNAL, "internal error")
			},
		}

		req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/contract/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestContract_ContractMakeRevision(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ContractService: &mock.ContractService{
				MakeRevisionFn: func(ctx context.Context, revision *entity.Revision) error {
					revision.CreatedAt = common.AppNowUTC()
					revision.Rev = 1
					if err := revision.Validate(); err != nil {
						return err
					}
					revision.ID = 1
					return nil
				},
			},
		}

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		file := mustOpen("revision_example/revision_test.js")
		defer file.Close()

		part, err := writer.CreateFormFile("compiled_revision", file.Name())
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(part, file)

		part, err = writer.CreateFormField("revision")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(revisionRequestBody)); err != nil {
			t.Fatal(err)
		}

		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/revision", &b)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var revisionData revisionResponse

		if err := json.NewDecoder(resp.Body).Decode(&revisionData); err != nil {
			t.Fatal(err)
		} else if revisionData.Revision == nil {
			t.Fatalf("expected contract data, got %v", revisionData)
		} else if revisionData.Revision.ID != 1 {
			t.Fatalf("expected revision id 1, got %d", revisionData.Revision.ID)
		} else if revisionData.Revision.ContractID != 1 {
			t.Fatalf("expected revision contract id 1, got %d", revisionData.Revision.ContractID)
		}
	})

	t.Run("InvalidContractID", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/invalid/revision", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("MissingJsonBody", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/revision", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("InvalidJsonBody", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)
		part, err := writer.CreateFormField("revision")
		if err != nil {
			t.Fatal(err)
		} else if _, err := io.WriteString(part, "invalid json"); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/revision", &b)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("MissingFile", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)
		part, err := writer.CreateFormField("revision")
		if err != nil {
			t.Fatal(err)
		} else if _, err := io.WriteString(part, revisionRequestBody); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/revision", &b)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ErrMakeRevision", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ContractService: &mock.ContractService{
				MakeRevisionFn: func(ctx context.Context, revision *entity.Revision) error {
					return apperr.Errorf(apperr.EINTERNAL, "internal error")
				},
			},
		}

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		file := mustOpen("revision_example/revision_test.js")
		defer file.Close()

		part, err := writer.CreateFormFile("compiled_revision", file.Name())
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(part, file)

		part, err = writer.CreateFormField("revision")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(revisionRequestBody)); err != nil {
			t.Fatal(err)
		}

		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/revision", &b)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestContract_ContractCall(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindContractByIDFn: func(ctx context.Context, id int64) (*entity.Contract, error) {
				if id == 1 {
					return &entity.Contract{
						ID:           1,
						Name:         "contract",
						Description:  "contract description",
						LastRevision: &entity.Revision{ID: 1},
						User: &entity.User{
							ID: 1,
						},
					}, nil
				}
				return nil, apperr.Errorf(apperr.ENOTFOUND, "contract not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ExecutorService: &mock.ExecutorService{
				ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
					return "OK", nil
				},
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/call", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		ContractCallResult := make(map[string]interface{})

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		} else if err := json.NewDecoder(resp.Body).Decode(&ContractCallResult); err != nil {
			t.Fatal(err)
		} else if ContractCallResult["result"] != "OK" {
			t.Fatalf("expected result %s, got %s", "OK", ContractCallResult["result"])
		}
	})

	t.Run("InvalidContractID", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindContractByIDFn: func(ctx context.Context, id int64) (*entity.Contract, error) {
				if id == 1 {
					return &entity.Contract{
						ID:           1,
						Name:         "contract",
						Description:  "contract description",
						LastRevision: &entity.Revision{ID: 1},
						User: &entity.User{
							ID: 1,
						},
					}, nil
				}
				return nil, apperr.Errorf(apperr.ENOTFOUND, "contract not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ExecutorService: &mock.ExecutorService{
				ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
					return "OK", nil
				},
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/invalid/call", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ContractNotFound", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindContractByIDFn: func(ctx context.Context, id int64) (*entity.Contract, error) {
				return nil, apperr.Errorf(apperr.ENOTFOUND, "contract not found")
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/call", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("ErrExecContract", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindContractByIDFn: func(ctx context.Context, id int64) (*entity.Contract, error) {
				if id == 1 {
					return &entity.Contract{
						ID:           1,
						Name:         "contract",
						Description:  "contract description",
						LastRevision: &entity.Revision{ID: 1},
						User: &entity.User{
							ID: 1,
						},
					}, nil
				}
				return nil, apperr.Errorf(apperr.ENOTFOUND, "contract not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ExecutorService: &mock.ExecutorService{
				ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
					return "", apperr.Errorf(apperr.EINTERNAL, "internal error")
				},
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/call", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestContract_ContractCallRev(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindRevisionByContractAndRevFn: func(ctx context.Context, contractID int64, rev entity.RevisionNumber) (*entity.Revision, error) {
				if contractID == 1 && rev == 1 {
					return &entity.Revision{
						ID:         1,
						Rev:        1,
						Version:    entity.CurrentRevisionVersion,
						ContractID: contractID,
					}, nil
				}
				return nil, apperr.Errorf(apperr.ENOTFOUND, "contract not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ExecutorService: &mock.ExecutorService{
				ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
					return "OK", nil
				},
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/call/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		ContractCallResult := make(map[string]interface{})

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		} else if err := json.NewDecoder(resp.Body).Decode(&ContractCallResult); err != nil {
			t.Fatal(err)
		} else if ContractCallResult["result"] != "OK" {
			t.Fatalf("expected result %s, got %s", "OK", ContractCallResult["result"])
		}
	})

	t.Run("InvalidContractID", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindContractByIDFn: func(ctx context.Context, id int64) (*entity.Contract, error) {
				if id == 1 {
					return &entity.Contract{
						ID:           1,
						Name:         "contract",
						Description:  "contract description",
						LastRevision: &entity.Revision{ID: 1},
						User: &entity.User{
							ID: 1,
						},
					}, nil
				}
				return nil, apperr.Errorf(apperr.ENOTFOUND, "contract not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ExecutorService: &mock.ExecutorService{
				ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
					return "OK", nil
				},
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/invalid/call/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("InvalidContractRev", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindContractByIDFn: func(ctx context.Context, id int64) (*entity.Contract, error) {
				if id == 1 {
					return &entity.Contract{
						ID:           1,
						Name:         "contract",
						Description:  "contract description",
						LastRevision: &entity.Revision{ID: 1},
						User: &entity.User{
							ID: 1,
						},
					}, nil
				}
				return nil, apperr.Errorf(apperr.ENOTFOUND, "contract not found")
			},
		}

		s.ServiceHandler.VmCallableService = &mock.VmCallableService{
			ExecutorService: &mock.ExecutorService{
				ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
					return "OK", nil
				},
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/call/invalid", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ContractRevNotFound", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		s.ServiceHandler.JWTService = &mock.JWTService{
			ParseFn: func(ctx context.Context, token string) (*entity.AppClaims, error) {
				if token == "OK" {
					return &entity.AppClaims{
						Auth: &entity.Auth{
							UserID: 1,
							ID:     1,
							User:   &entity.User{ID: 1},
						},
					}, nil
				}

				return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")
			},
		}

		s.ServiceHandler.UserSearchService = &mock.UserService{
			FindUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				if id == 1 {
					return &entity.User{ID: 1}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "user not found")
			},
		}

		s.ServiceHandler.AuthSearchService = &mock.AuthService{
			FindAuthByIDFn: func(ctx context.Context, id int64) (*entity.Auth, error) {
				if id == 1 {
					return &entity.Auth{
						UserID: 1,
						ID:     1,
						User:   &entity.User{ID: 1},
					}, nil
				}

				return nil, apperr.Errorf(apperr.ENOTFOUND, "auth not found")
			},
		}

		s.ServiceHandler.ContractSearchService = &mock.ContractService{
			FindRevisionByContractAndRevFn: func(ctx context.Context, contractID int64, rev entity.RevisionNumber) (*entity.Revision, error) {
				return nil, apperr.Errorf(apperr.ENOTFOUND, "contract not found")
			},
		}

		req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/contract/1/call/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer OK")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

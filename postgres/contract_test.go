package postgres_test

import (
	"context"
	"testing"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/postgres"
)

func TestContract_CreateContract(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		user, ctx := MustCreateUser(t, context.Background(), db, &entity.User{Name: "test-contract"})

		testContractName := "test-contract"
		testContractDescription := "test-contract-description"

		contract := &entity.Contract{
			Name:        testContractName,
			Description: testContractDescription,
			UserID:      user.ID,
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		if err := cs.CreateContract(ctx, contract); err != nil {
			t.Fatal(err)
		} else if contract.ID == 0 {
			t.Fatal("contract ID is zero")
		} else if contract.CreatedAt.IsZero() {
			t.Fatal("contract created at is zero")
		} else if contract.UpdatedAt.IsZero() {
			t.Fatal("contract updated at is zero")
		} else if contract.User == nil {
			t.Fatal("contract user is nil")
		} else if contract.User.ID != user.ID {
			t.Fatalf("contract user ID is %d, expected %d", contract.User.ID, user.ID)
		} else if contract.Name != testContractName {
			t.Fatalf("contract name is %s, expected %s", contract.Name, testContractName)
		} else if contract.Description != testContractDescription {
			t.Fatalf("contract description is %s, expected %s", contract.Description, testContractDescription)
		} else if contract.Visibility != entity.VisibilityPublic {
			t.Fatalf("contract visibility is %s, expected %s", contract.Visibility, entity.VisibilityPublic)
		} else if contract.MaxFuel != entity.FuelExtremeActionAmount {
			t.Fatalf("contract max fuel is %d, expected %d", contract.MaxFuel, entity.FuelExtremeActionAmount)
		}
	})

	t.Run("ContextCancelled", func(t *testing.T) {

		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		user, ctx := MustCreateUser(t, context.Background(), db, &entity.User{Name: "test-contract"})

		testContractName := "test-contract"
		testContractDescription := "test-contract-description"

		contract := &entity.Contract{
			Name:        testContractName,
			Description: testContractDescription,
			UserID:      user.ID,
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		ctx, cancel := context.WithCancel(ctx)

		cancel()

		if err := cs.CreateContract(ctx, contract); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ErrValidate", func(t *testing.T) {

		t.Run("MissingName", func(t *testing.T) {
			db := MustOpenDB(t)
			defer MustCloseDB(t, db)

			TruncateTablesForContractTests(t, db)

			cs := postgres.NewContractService(db)

			user, ctx := MustCreateUser(t, context.Background(), db, &entity.User{Name: "test-contract"})

			testContractDescription := "test-contract-description"

			contract := &entity.Contract{
				Description: testContractDescription,
				UserID:      user.ID,
				Visibility:  entity.VisibilityPublic,
				MaxFuel:     entity.FuelExtremeActionAmount,
			}

			if err := cs.CreateContract(ctx, contract); err == nil {
				t.Fatal("expected error")
			} else if code := apperr.ErrorCode(err); code != apperr.EINVALID {
				t.Fatalf("expected EINVALID, got %s", code)
			}
		})

		t.Run("MissingUserID", func(t *testing.T) {
			db := MustOpenDB(t)
			defer MustCloseDB(t, db)

			TruncateTablesForContractTests(t, db)

			cs := postgres.NewContractService(db)

			testContractName := "test-contract"
			testContractDescription := "test-contract-description"

			contract := &entity.Contract{
				Name:        testContractName,
				Description: testContractDescription,
				Visibility:  entity.VisibilityPublic,
				MaxFuel:     entity.FuelExtremeActionAmount,
			}

			if err := cs.CreateContract(context.Background(), contract); err == nil {
				t.Fatal("expected error")
			} else if code := apperr.ErrorCode(err); code != apperr.EINVALID {
				t.Fatalf("expected EINVALID, got %s", code)
			}
		})

		t.Run("MissingOrWrongVisibility", func(t *testing.T) {

			db := MustOpenDB(t)
			defer MustCloseDB(t, db)

			TruncateTablesForContractTests(t, db)

			cs := postgres.NewContractService(db)

			user, ctx := MustCreateUser(t, context.Background(), db, &entity.User{Name: "test-contract"})

			testContractName := "test-contract"
			testContractDescription := "test-contract-description"

			contract := &entity.Contract{
				Name:        testContractName,
				Description: testContractDescription,
				UserID:      user.ID,
				Visibility:  "wrong",
				MaxFuel:     entity.FuelExtremeActionAmount,
			}

			if err := cs.CreateContract(ctx, contract); err == nil {
				t.Fatal("expected error")
			} else if code := apperr.ErrorCode(err); code != apperr.EINVALID {
				t.Fatalf("expected EINVALID, got %s", code)
			}

			contract = &entity.Contract{
				Name:        testContractName,
				Description: testContractDescription,
				UserID:      user.ID,
				MaxFuel:     entity.FuelExtremeActionAmount,
			}

			if err := cs.CreateContract(ctx, contract); err == nil {
				t.Fatal("expected error")
			} else if code := apperr.ErrorCode(err); code != apperr.EINVALID {
				t.Fatalf("expected EINVALID, got %s", code)
			}
		})
	})
}

func MustCreateContract(tb testing.TB, ctx context.Context, db *postgres.DB, contract *entity.Contract) (*entity.Contract, context.Context) {
	tb.Helper()
	if err := postgres.NewContractService(db).CreateContract(ctx, contract); err != nil {
		tb.Fatal(err)
	}
	return contract, app.NewContextWithUser(ctx, contract.User)
}

func TruncateTablesForContractTests(tb testing.TB, db *postgres.DB) {
	tb.Helper()

	MustTruncateTable(tb, db, "users")
	MustTruncateTable(tb, db, "contracts")
}

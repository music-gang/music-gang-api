package postgres_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
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

	t.Run("UserNotAuthenticated", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		user, _ := MustCreateUser(t, context.Background(), db, &entity.User{Name: "test-contract"})

		testContractName := "test-contract"
		testContractDescription := "test-contract-description"

		contract := &entity.Contract{
			Name:        testContractName,
			Description: testContractDescription,
			UserID:      user.ID,
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		if err := cs.CreateContract(context.Background(), contract); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.EUNAUTHORIZED {
			t.Fatalf("expected %s, got %s", apperr.EUNAUTHORIZED, code)
		}
	})

	t.Run("UserNotContractOwner", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		user0, _ := MustCreateUser(t, context.Background(), db, &entity.User{Name: "test-contract"})

		_, ctx1 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "test-contract-2"})

		testContractName := "test-contract"
		testContractDescription := "test-contract-description"

		contract := &entity.Contract{
			Name:        testContractName,
			Description: testContractDescription,
			UserID:      user0.ID,
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		if err := cs.CreateContract(ctx1, contract); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.EUNAUTHORIZED {
			t.Fatalf("expected %s, got %s", apperr.EUNAUTHORIZED, code)
		}
	})

	t.Run("UserDeletedBeforeCreateContract", func(t *testing.T) {

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

		if err := postgres.NewUserService(db).DeleteUser(ctx, user.ID); err != nil {
			t.Fatal(err)
		} else if err := cs.CreateContract(ctx, contract); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, code)
		}
	})
}

func TestContract_DeleteContract(t *testing.T) {

	userForContract := &entity.User{Name: "test-contract"}

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		cs := postgres.NewContractService(db)

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, userForContract)

		if err := cs.DeleteContract(ctx, contract.ID); err != nil {
			t.Fatal(err)
		} else if _, err := cs.FindContractByID(ctx, contract.ID); err == nil {
			t.Fatalf("expected error, got %v", err)
		} else if code := apperr.ErrorCode(err); code != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		cs := postgres.NewContractService(db)

		if err := cs.DeleteContract(context.Background(), contract.ID); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, code)
		}
	})

	t.Run("NotOwned", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		cs := postgres.NewContractService(db)

		contract0, _ := MustCreateContract(t, context.Background(), db, contract, userForContract)

		_, ctx1 := MustCreateUser(t, context.Background(), db, &entity.User{Name: "test-contract-2"})

		if err := cs.DeleteContract(ctx1, contract0.ID); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.EUNAUTHORIZED {
			t.Fatalf("expected %s, got %s", apperr.EUNAUTHORIZED, code)
		}
	})

	t.Run("UserDeletedBeforeDeleteContract", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		cs := postgres.NewContractService(db)

		contract0, ctx := MustCreateContract(t, context.Background(), db, contract, userForContract)

		if err := postgres.NewUserService(db).DeleteUser(ctx, contract.UserID); err != nil {
			t.Fatal(err)
		} else if err := cs.DeleteContract(ctx, contract0.ID); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, code)
		}
	})
}

func TestContract_UpdateContract(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-contract"})

		oldUpdatedAt := contract.UpdatedAt
		newContractName := "new-test-contract"
		newContractDescription := "new-test-contract-description"

		contractUpdate := service.ContractUpdate{
			Name:        &newContractName,
			Description: &newContractDescription,
		}

		contract.Name = newContractName
		contract.Description = newContractDescription

		// sleep to make sure updated_at is different
		time.Sleep(time.Second)

		if c, err := cs.UpdateContract(ctx, contract.ID, contractUpdate); err != nil {
			t.Fatal(err)
		} else if err := CompareContracts(t, contract, c); err != nil {
			t.Fatal(err)
		} else if c.UpdatedAt.Equal(oldUpdatedAt) {
			t.Fatalf("expected updatedAt to be updated")
		} else if reloadedContract, err := cs.FindContractByID(ctx, c.ID); err != nil {
			t.Fatal(err)
		} else if err := CompareContracts(t, reloadedContract, c); err != nil {
			t.Fatal(err)

		} else if reloadedContract.UpdatedAt.Equal(oldUpdatedAt) {
			t.Fatalf("expected updatedAt to be updated")
		} else if !reloadedContract.UpdatedAt.Equal(c.UpdatedAt) {
			t.Fatalf("expected updatedAt to be equal")
		}
	})

	t.Run("NotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-contract"})

		if err := cs.DeleteContract(ctx, contract.ID); err != nil {
			t.Fatal(err)
		}

		newContractName := "new-test-contract"
		newContractDescription := "new-test-contract-description"

		contractUpdate := service.ContractUpdate{
			Name:        &newContractName,
			Description: &newContractDescription,
		}

		if _, err := cs.UpdateContract(ctx, contract.ID, contractUpdate); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, code)
		}
	})

	t.Run("NotOwned", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		contract, _ = MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-contract"})

		newContractName := "new-test-contract"
		newContractDescription := "new-test-contract-description"

		contractUpdate := service.ContractUpdate{
			Name:        &newContractName,
			Description: &newContractDescription,
		}

		_, ctx := MustCreateUser(t, context.Background(), db, &entity.User{Name: "test-contract-another-user"})

		if _, err := cs.UpdateContract(ctx, contract.ID, contractUpdate); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.EUNAUTHORIZED {
			t.Fatalf("expected %s, got %s", apperr.EUNAUTHORIZED, code)
		}
	})

	t.Run("NotValid", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-contract"})

		newContractName := ""
		newContractDescription := "new-test-contract-description"

		contractUpdate := service.ContractUpdate{
			Name:        &newContractName,
			Description: &newContractDescription,
		}

		if _, err := cs.UpdateContract(ctx, contract.ID, contractUpdate); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.EINVALID {
			t.Fatalf("expected %s, got %s", apperr.EINVALID, code)
		}
	})

	t.Run("UserDeletedBeforeUpdateContract", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-contract"})

		newContractName := "new-test-contract"
		newContractDescription := "new-test-contract-description"

		contractUpdate := service.ContractUpdate{
			Name:        &newContractName,
			Description: &newContractDescription,
		}

		if err := postgres.NewUserService(db).DeleteUser(ctx, contract.UserID); err != nil {
			t.Fatal(err)
		}

		if _, err := cs.UpdateContract(ctx, contract.ID, contractUpdate); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, code)
		}
	})
}

func TestContract_FindContractByID(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		userContract := &entity.User{Name: "test-contract"}

		contract := &entity.Contract{
			Name:        "test-contract",
			Description: "test-contract-description",
			Visibility:  entity.VisibilityPublic,
			MaxFuel:     entity.FuelExtremeActionAmount,
		}

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, userContract)

		if contract.ID == 0 {
			t.Fatal("expected non-zero contract id")
		}

		if c, err := cs.FindContractByID(ctx, contract.ID); err != nil {
			t.Fatal(err)
		} else if err := CompareContracts(t, contract, c); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		if _, err := cs.FindContractByID(context.Background(), 1); err == nil {
			t.Fatal("expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, code)
		}
	})
}

func TestContract_FindContracts(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		for i := 0; i < 10; i++ {

			tempUserContract := &entity.User{Name: "test-contract-" + strconv.Itoa(i)}

			contract := &entity.Contract{
				Name:        "test-contract",
				Description: "test-contract-description",
				Visibility:  entity.VisibilityPublic,
				MaxFuel:     entity.FuelExtremeActionAmount,
			}

			MustCreateContract(t, context.Background(), db, contract, tempUserContract)
		}

		filterName := "test-contract"
		filterDescription := "test-contract-description"

		if contracts, n, err := cs.FindContracts(context.Background(), service.ContractFilter{
			Name:        &filterName,
			Description: &filterDescription,
		}); err != nil {
			t.Fatal(err)
		} else if n != 10 {
			t.Fatalf("expected 10, got %d", n)
		} else if len(contracts) != n {
			t.Fatalf("expected %d, got %d", n, len(contracts))
		}

		if contracts, n, err := cs.FindContracts(context.Background(), service.ContractFilter{Limit: 2}); err != nil {
			t.Fatal(err)
		} else if n != 10 {
			t.Fatalf("expected 10, got %d", n)
		} else if len(contracts) == n {
			t.Fatal("Total number of contracts should be greater than the limit")
		} else if len(contracts) > 2 {
			t.Fatal("Expected only 2 contracts")
		}

		if contracts, n, err := cs.FindContracts(context.Background(), service.ContractFilter{Offset: 2, Limit: 2}); err != nil {
			t.Fatal(err)
		} else if n != 10 {
			t.Fatalf("expected 10, got %d", n)
		} else if len(contracts) == n {
			t.Fatal("Total number of contracts should be greater than the limit")
		} else if len(contracts) > 2 {
			t.Fatal("Expected only 2 contracts")
		}

		if contracts, n, err := cs.FindContracts(context.Background(), service.ContractFilter{Offset: 10, Limit: 2}); err != nil {
			t.Fatal(err)
		} else if n != 0 {
			t.Fatalf("expected 10, got %d", n)
		} else if len(contracts) != 0 {
			t.Fatal("Expected no contracts")
		}
	})
}

func MustCreateContract(tb testing.TB, ctx context.Context, db *postgres.DB, contract *entity.Contract, user *entity.User) (*entity.Contract, context.Context) {
	tb.Helper()
	user, ctx = MustCreateUser(tb, ctx, db, user)
	contract.UserID = user.ID
	if err := postgres.NewContractService(db).CreateContract(ctx, contract); err != nil {
		tb.Fatal(err)
	}
	return contract, ctx
}

func TruncateTablesForContractTests(tb testing.TB, db *postgres.DB) {
	tb.Helper()

	MustTruncateTable(tb, db, "users")
	MustTruncateTable(tb, db, "contracts")
}

func CompareContracts(t testing.TB, expected *entity.Contract, actual *entity.Contract) error {
	t.Helper()

	if expected.ID != actual.ID {
		return fmt.Errorf("expected contract id %d, got %d", expected.ID, actual.ID)
	}

	if expected.Name != actual.Name {
		return fmt.Errorf("expected contract name %s, got %s", expected.Name, actual.Name)
	}

	if expected.Description != actual.Description {
		return fmt.Errorf("expected contract description %s, got %s", expected.Description, actual.Description)
	}

	if expected.Visibility != actual.Visibility {
		return fmt.Errorf("expected contract visibility %s, got %s", expected.Visibility, actual.Visibility)
	}

	if expected.MaxFuel != actual.MaxFuel {
		return fmt.Errorf("expected contract max fuel %d, got %d", expected.MaxFuel, actual.MaxFuel)
	}

	if expected.UserID != actual.UserID {
		return fmt.Errorf("expected contract user id %d, got %d", expected.UserID, actual.UserID)
	}

	return nil
}

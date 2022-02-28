package postgres_test

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
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
		}

		if other, err := cs.FindContractByID(ctx, contract.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(contract, other) {
			t.Fatal("contracts are not equal")
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

		t.Run("InvalidFuel", func(t *testing.T) {

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

		// sleep to make sure updated_at is different
		time.Sleep(time.Second)

		if c, err := cs.UpdateContract(ctx, contract.ID, contractUpdate); err != nil {
			t.Fatal(err)
		} else if cc, err := cs.FindContractByID(ctx, c.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(c, cc) {
			t.Fatal("expected contract to be equal")
		} else if c.UpdatedAt.Equal(oldUpdatedAt) {
			t.Fatalf("expected updatedAt to be updated")
		} else if reloadedContract, err := cs.FindContractByID(ctx, c.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(reloadedContract, c) {
			t.Fatal("expected contract to be equal")
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
		} else if !reflect.DeepEqual(contract, c) {
			t.Fatal("expected contract to be equal")
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

func TestContract_MakeRevision(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:       "make-revision",
			MaxFuel:    entity.FuelExtremeActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-make-revision"})

		code := "test-code"
		notes := "test-notes"

		revision := &entity.Revision{
			Version:      entity.AnchorageVersion,
			ContractID:   contract.ID,
			Code:         code,
			Notes:        notes,
			CompiledCode: []byte(code),
			MaxFuel:      entity.FuelInstantActionAmount,
		}

		if err := cs.MakeRevision(ctx, revision); err != nil {
			t.Fatal(err)
		} else if r, err := cs.FindRevisionByContractAndRev(ctx, contract.ID, revision.Rev); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(revision, r) {
			t.Fatal("expected revision to be equal")
		}
	})

	t.Run("MaxFuelDefaultFromContract", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:       "make-revision",
			MaxFuel:    entity.FuelExtremeActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-make-revision"})

		code := "test-code"
		notes := "test-notes"

		revision := &entity.Revision{
			Version:      entity.AnchorageVersion,
			ContractID:   contract.ID,
			Code:         code,
			Notes:        notes,
			CompiledCode: []byte(code),
		}

		if err := cs.MakeRevision(ctx, revision); err != nil {
			t.Fatal(err)
		} else if revision.MaxFuel != contract.MaxFuel {
			t.Fatalf("expected revision max fuel to be %d, got %d", contract.MaxFuel, revision.MaxFuel)
		}
	})

	t.Run("NewRevision", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:       "make-revision",
			MaxFuel:    entity.FuelExtremeActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-make-revision"})

		code := "test-code"
		notes := "test-notes"

		revision := &entity.Revision{
			Version:      entity.AnchorageVersion,
			ContractID:   contract.ID,
			Code:         code,
			Notes:        notes,
			CompiledCode: []byte(code),
			MaxFuel:      entity.FuelInstantActionAmount,
		}

		if err := cs.MakeRevision(ctx, revision); err != nil {
			t.Fatal(err)
		} else if revision.Rev != 1 {
			t.Fatalf("expected revision rev to be 1, got %d", revision.Rev)
		}

		newCode := "test-code"
		newNotes := "test-notes"

		newRevision := &entity.Revision{
			Version:      entity.AnchorageVersion,
			ContractID:   contract.ID,
			Code:         newCode,
			Notes:        newNotes,
			CompiledCode: []byte(newCode),
			MaxFuel:      entity.FuelInstantActionAmount,
		}

		if err := cs.MakeRevision(ctx, newRevision); err != nil {
			t.Fatal(err)
		} else if newRevision.Rev != 2 {
			t.Fatalf("expected revision rev to be 2, got %d", newRevision.Rev)
		}
	})

	t.Run("ContractNotFound", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		code := "test-code"
		notes := "test-notes"

		revision := &entity.Revision{
			Version:      entity.AnchorageVersion,
			ContractID:   1,
			Code:         code,
			Notes:        notes,
			CompiledCode: []byte(code),
			MaxFuel:      entity.FuelInstantActionAmount,
		}

		if err := cs.MakeRevision(context.Background(), revision); err == nil {
			t.Fatal(err)
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.ENOTFOUND {
			t.Fatalf("expected error code to be %s, got %s", apperr.ENOTFOUND, errCode)
		}
	})

	t.Run("ContractNotOwned", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract, _ := MustCreateContract(t, context.Background(), db, &entity.Contract{
			Name:       "make-revision",
			MaxFuel:    entity.FuelExtremeActionAmount,
			Visibility: entity.VisibilityPublic,
		}, &entity.User{Name: "test-make-revision"})

		_, ctx1 := MustCreateContract(t, context.Background(), db, &entity.Contract{
			Name:       "make-revision",
			MaxFuel:    entity.FuelExtremeActionAmount,
			Visibility: entity.VisibilityPublic,
		}, &entity.User{Name: "test-make-revision-1"})

		code := "test-code"
		notes := "test-notes"

		revision := &entity.Revision{
			Version:      entity.AnchorageVersion,
			ContractID:   contract.ID,
			Code:         code,
			Notes:        notes,
			CompiledCode: []byte(code),
			MaxFuel:      entity.FuelInstantActionAmount,
		}

		if err := cs.MakeRevision(ctx1, revision); err == nil {
			t.Fatal(err)
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EUNAUTHORIZED {
			t.Fatalf("expected error code to be %s, got %s", apperr.EUNAUTHORIZED, errCode)
		}
	})

	t.Run("ContextCancelled", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:       "make-revision",
			MaxFuel:    entity.FuelExtremeActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-make-revision"})

		code := "test-code"
		notes := "test-notes"

		revision := &entity.Revision{
			Version:      entity.AnchorageVersion,
			ContractID:   contract.ID,
			Code:         code,
			Notes:        notes,
			CompiledCode: []byte(code),
			MaxFuel:      entity.FuelInstantActionAmount,
		}

		ctx, cancel := context.WithCancel(ctx)

		cancel()

		if err := cs.MakeRevision(ctx, revision); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})

	t.Run("ErrValidate", func(t *testing.T) {

		contract := &entity.Contract{
			Name:       "make-revision",
			MaxFuel:    entity.FuelExtremeActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		t.Run("MissingVersion", func(t *testing.T) {

			db := MustOpenDB(t)
			defer db.Close()

			TruncateTablesForContractTests(t, db)

			cs := postgres.NewContractService(db)

			contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-make-revision"})

			code := "test-code"
			notes := "test-notes"

			revision := &entity.Revision{
				ContractID:   contract.ID,
				Code:         code,
				Notes:        notes,
				CompiledCode: []byte(code),
				MaxFuel:      entity.FuelInstantActionAmount,
			}

			if err := cs.MakeRevision(ctx, revision); err == nil {
				t.Fatal("expected error")
			} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINVALID {
				t.Fatalf("expected %s, got %s", apperr.EINVALID, errCode)
			}
		})

		t.Run("MissingContractID", func(t *testing.T) {

			db := MustOpenDB(t)
			defer db.Close()

			TruncateTablesForContractTests(t, db)

			cs := postgres.NewContractService(db)

			code := "test-code"
			notes := "test-notes"

			revision := &entity.Revision{
				Version:      entity.AnchorageVersion,
				Code:         code,
				Notes:        notes,
				CompiledCode: []byte(code),
				MaxFuel:      entity.FuelInstantActionAmount,
			}

			if err := cs.MakeRevision(context.Background(), revision); err == nil {
				t.Fatal("expected error")
			} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINVALID {
				t.Fatalf("expected %s, got %s", apperr.EINVALID, errCode)
			}
		})

		t.Run("MissingCode", func(t *testing.T) {

			db := MustOpenDB(t)
			defer db.Close()

			TruncateTablesForContractTests(t, db)

			cs := postgres.NewContractService(db)

			contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-make-revision"})

			code := "test-code"
			notes := "test-notes"

			revision := &entity.Revision{
				Version:      entity.AnchorageVersion,
				ContractID:   contract.ID,
				Notes:        notes,
				CompiledCode: []byte(code),
				MaxFuel:      entity.FuelInstantActionAmount,
			}

			if err := cs.MakeRevision(ctx, revision); err == nil {
				t.Fatal("expected error")
			} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINVALID {
				t.Fatalf("expected %s, got %s", apperr.EINVALID, errCode)
			}
		})

		t.Run("MissingCompiledCode", func(t *testing.T) {

			db := MustOpenDB(t)
			defer db.Close()

			TruncateTablesForContractTests(t, db)

			cs := postgres.NewContractService(db)

			contract, ctx := MustCreateContract(t, context.Background(), db, contract, &entity.User{Name: "test-make-revision"})

			code := "test-code"
			notes := "test-notes"

			revision := &entity.Revision{
				Version:    entity.AnchorageVersion,
				ContractID: contract.ID,
				Notes:      notes,
				Code:       code,
				MaxFuel:    entity.FuelInstantActionAmount,
			}

			if err := cs.MakeRevision(ctx, revision); err == nil {
				t.Fatal("expected error")
			} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINVALID {
				t.Fatalf("expected %s, got %s", apperr.EINVALID, errCode)
			}
		})
	})
}

func TestContract_FindRevisionByContractAndRev(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:       "test-find-revision-by-contract-and-rev",
			MaxFuel:    entity.FuelInstantActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		revision, ctx := MustCreateRevision(t, context.Background(), db, DataToMakeRevision{
			Contract: contract,
			User:     &entity.User{Name: "test-find-revision-by-contract-and-rev"},
			Revision: &entity.Revision{
				Code:         "test-code",
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
		})

		if r, err := cs.FindRevisionByContractAndRev(ctx, revision.ContractID, revision.Rev); err != nil {
			t.Fatal(err)
		} else if r == nil {
			t.Fatal("expected revision")
		} else if !reflect.DeepEqual(r, revision) {
			t.Fatalf("expected revision %v, got %v", revision, r)
		}
	})

	t.Run("ContextCancelled", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:       "test-find-revision-by-contract-and-rev",
			MaxFuel:    entity.FuelInstantActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		_, ctx := MustCreateRevision(t, context.Background(), db, DataToMakeRevision{
			Contract: contract,
			User:     &entity.User{Name: "test-find-revision-by-contract-and-rev"},
			Revision: &entity.Revision{
				Code:         "test-code",
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
		})

		ctx, cancel := context.WithCancel(ctx)

		cancel()

		if _, err := cs.FindRevisionByContractAndRev(ctx, contract.ID, 0); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("expected %s, got %s", apperr.EINTERNAL, errCode)
		}
	})

	t.Run("ContractNotExists", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		if _, err := cs.FindRevisionByContractAndRev(context.Background(), 1, 0); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, errCode)
		}
	})

	t.Run("RevNumberNotExists", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:       "test-find-revision-by-contract-and-rev",
			MaxFuel:    entity.FuelInstantActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		MustCreateRevision(t, context.Background(), db, DataToMakeRevision{
			Contract: contract,
			User:     &entity.User{Name: "test-find-revision-by-contract-and-rev"},
			Revision: &entity.Revision{
				Code:         "test-code",
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
		})

		// now exists rev n°1, try to find rev n°2

		if _, err := cs.FindRevisionByContractAndRev(context.Background(), contract.ID, 2); err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.ENOTFOUND {
			t.Fatalf("expected %s, got %s", apperr.ENOTFOUND, errCode)
		}
	})

	t.Run("FindLastRevision", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		TruncateTablesForContractTests(t, db)

		cs := postgres.NewContractService(db)

		contract := &entity.Contract{
			Name:       "test-find-revision-by-contract-and-rev",
			MaxFuel:    entity.FuelInstantActionAmount,
			Visibility: entity.VisibilityPublic,
		}

		revision, ctx := MustCreateRevision(t, context.Background(), db, DataToMakeRevision{
			Contract: contract,
			User:     &entity.User{Name: "test-find-revision-by-contract-and-rev"},
			Revision: &entity.Revision{
				Code:         "test-code",
				CompiledCode: []byte("test-code"),
				Version:      entity.CurrentRevisionVersion,
			},
		})

		// create random number between 2 and 100
		rand.Seed(time.Now().UnixNano())
		num := rand.Intn(98) + 2

		for i := 0; i < num; i++ {
			revisionCode := fmt.Sprintf("test-code-%d", i)
			revision := &entity.Revision{
				Version:      entity.AnchorageVersion,
				ContractID:   contract.ID,
				Code:         revisionCode,
				CompiledCode: []byte(revisionCode),
				MaxFuel:      entity.FuelInstantActionAmount,
			}
			if err := cs.MakeRevision(ctx, revision); err != nil {
				t.Fatal(err)
			}
		}

		lastRevision, err := cs.FindRevisionByContractAndRev(ctx, revision.ContractID, 0)
		if err != nil {
			t.Fatal(err)
		}

		lastCalculatedRevision, err := cs.FindRevisionByContractAndRev(ctx, revision.ContractID, entity.RevisionNumber(num+1))
		if err != nil {
			t.Fatal(err)
		}

		if lastRevision.Rev != lastCalculatedRevision.Rev {
			t.Fatalf("expected last revision %d, got %d", lastCalculatedRevision.Rev, lastRevision.Rev)
		} else if !reflect.DeepEqual(lastRevision, lastCalculatedRevision) {
			t.Fatalf("expected last revision %v, got %v", lastCalculatedRevision, lastRevision)
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

type DataToMakeRevision struct {
	Contract *entity.Contract
	Revision *entity.Revision
	User     *entity.User
}

func MustCreateRevision(tb testing.TB, ctx context.Context, db *postgres.DB, data DataToMakeRevision) (*entity.Revision, context.Context) {
	tb.Helper()
	contract, ctx := MustCreateContract(tb, ctx, db, data.Contract, data.User)
	data.Revision.ContractID = contract.ID
	if err := postgres.NewContractService(db).MakeRevision(ctx, data.Revision); err != nil {
		tb.Fatal(err)
	}
	return data.Revision, ctx
}

func TruncateTablesForContractTests(tb testing.TB, db *postgres.DB) {
	tb.Helper()

	MustTruncateTable(tb, db, "users")
	MustTruncateTable(tb, db, "auths")
	MustTruncateTable(tb, db, "contracts")
}

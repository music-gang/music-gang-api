package mgvm_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/mgvm"
	"github.com/music-gang/music-gang-api/mock"
)

func TestFuelTank_Burn(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		currentFuel := entity.Fuel(0)

		fuelTank := mgvm.NewFuelTank()

		ctx := context.Background()

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if err := fuelTank.Burn(ctx, entity.Fuel(10)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if currentFuel != 10 {
			t.Fatalf("unexpected currentFuel: %v", currentFuel)
		}

		if err := fuelTank.Burn(ctx, entity.Fuel(30)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if currentFuel != 40 {
			t.Fatalf("unexpected currentFuel: %v", currentFuel)
		}
	})

	t.Run("MaxCapReached", func(t *testing.T) {

		currentFuel := entity.Fuel(0)

		fuelTank := mgvm.NewFuelTank()

		ctx := context.Background()

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if err := fuelTank.Burn(ctx, entity.Fuel(10)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if currentFuel != 10 {
			t.Fatalf("unexpected currentFuel: %v", currentFuel)
		}

		// error must by ErrFuelTankNotEnough

		if err := fuelTank.Burn(ctx, 10+entity.FuelTankCapacity); err == nil {
			t.Fatalf("expected error")
		} else if err != service.ErrFuelTankNotEnough {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("FuelErr", func(t *testing.T) {

		currentFuel := entity.Fuel(0)

		fuelTank := mgvm.NewFuelTank()

		ctx := context.Background()

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return 0, apperr.Errorf(apperr.EINTERNAL, "fuel-mock")
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if err := fuelTank.Burn(ctx, entity.Fuel(10)); err == nil {
			t.Fatalf("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("unexpected error: %v", err)
		} else if errMsg := apperr.ErrorMessage(err); errMsg != "fuel-mock" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("BurnErr", func(t *testing.T) {

		currentFuel := entity.Fuel(0)

		fuelTank := mgvm.NewFuelTank()

		ctx := context.Background()

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return apperr.Errorf(apperr.EINTERNAL, "burn-mock")
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if err := fuelTank.Burn(ctx, entity.Fuel(10)); err == nil {
			t.Fatalf("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("unexpected error: %v", err)
		} else if errMsg := apperr.ErrorMessage(err); errMsg != "burn-mock" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestFuelTank_Fuel(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		currentFuel := entity.Fuel(0)

		fuelTank := mgvm.NewFuelTank()

		ctx := context.Background()

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if cf, err := fuelTank.Fuel(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf != 0 {
			t.Fatalf("unexpected currentFuel: %v", cf)
		}
	})

	t.Run("LocalFuel", func(t *testing.T) {

		mgvm.SwitchToLocalFuel()
		defer mgvm.SwitchToRemoteFuel()

		currentFuel := entity.Fuel(0)

		fuelTank := mgvm.NewFuelTank()

		ctx := context.Background()

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if cf, err := fuelTank.Fuel(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf != 0 {
			t.Fatalf("unexpected currentFuel: %v", cf)
		}
	})
}

func TestFuelTank_Refuel(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		currentFuel := entity.Fuel(0)

		fuelTank := mgvm.NewFuelTank()

		ctx := context.Background()

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				currentFuel -= fuelToRefill
				return nil
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if err := fuelTank.Burn(ctx, entity.Fuel(50)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cf, err := fuelTank.Fuel(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf != 50 {
			t.Fatalf("unexpected currentFuel, got: %v, want: %v", cf, currentFuel)
		}

		if err := fuelTank.Refuel(ctx, entity.Fuel(30)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cf, err := fuelTank.Fuel(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf != 20 {
			t.Fatalf("unexpected currentFuel, got: %v, want: %v", cf, currentFuel)
		}

		// refuel over the limit

		if err := fuelTank.Refuel(ctx, entity.Fuel(30)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf, err := fuelTank.Fuel(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf != 0 {
			t.Fatalf("unexpected currentFuel, got: %v, want: %v", cf, currentFuel)
		}

		// local refuel

		mgvm.SwitchToLocalFuel()
		defer mgvm.SwitchToRemoteFuel()

		if err := fuelTank.Burn(ctx, entity.Fuel(50)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf, err := fuelTank.Fuel(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf != 50 {
			t.Fatalf("unexpected currentFuel, got: %v, want: %v", cf, currentFuel)
		} else if err := fuelTank.Refuel(ctx, entity.Fuel(30)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf, err := fuelTank.Fuel(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf != 20 {
			t.Fatalf("unexpected currentFuel, got: %v, want: %v", cf, currentFuel)
		}

		remoteFuel, err := mgvm.Fuel(ctx, fuelTank, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		localFuel, err := mgvm.Fuel(ctx, fuelTank, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if remoteFuel != localFuel {
			t.Fatalf("unexpected fuel, got: %v, want: %v", remoteFuel, localFuel)
		}
	})

	t.Run("RefuelErr", func(t *testing.T) {

		currentFuel := entity.Fuel(0)

		fuelTank := mgvm.NewFuelTank()

		ctx := context.Background()

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				return apperr.Errorf(apperr.EINTERNAL, "refuel-mock")
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if err := fuelTank.Burn(ctx, entity.Fuel(50)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cf, err := fuelTank.Fuel(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf != 50 {
			t.Fatalf("unexpected currentFuel, got: %v, want: %v", cf, currentFuel)
		}

		if err := fuelTank.Refuel(ctx, entity.Fuel(30)); err == nil {
			t.Fatalf("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("unexpected error code, got: %v, want: %v", errCode, apperr.EINTERNAL)
		} else if errMsg := apperr.ErrorMessage(err); errMsg != "refuel-mock" {
			t.Fatalf("unexpected error message, got: %v, want: %v", errMsg, "refuel-mock")
		}
	})

	t.Run("FuelErr", func(t *testing.T) {

		fuelTank := mgvm.NewFuelTank()

		currentFuel := entity.Fuel(0)

		ctx := context.Background()

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				currentFuel -= fuelToRefill
				return nil
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if err := fuelTank.Burn(ctx, entity.Fuel(50)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cf, err := fuelTank.Fuel(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if cf != 50 {
			t.Fatalf("unexpected currentFuel, got: %v, want: %v", cf, currentFuel)
		}

		fuelTank.FuelTankService = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, apperr.Errorf(apperr.EINTERNAL, "fuel-mock")
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				currentFuel -= fuelToRefill
				return nil
			},
		}

		if err := fuelTank.Refuel(ctx, entity.Fuel(30)); err == nil {
			t.Fatalf("expected error, got: %v", err)
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Fatalf("unexpected error code, got: %v, want: %v", errCode, apperr.EINTERNAL)
		} else if errMsg := apperr.ErrorMessage(err); errMsg != "fuel-mock" {
			t.Fatalf("unexpected error message, got: %v, want: %v", errMsg, "fuel-mock")
		}
	})
}

func TestFuelTank_Stats(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		fuelTank := mgvm.NewFuelTank()

		ctx := context.Background()

		currentFuel := entity.Fuel(0)

		now := time.Now()

		time.Sleep(1 * time.Second)

		fuelTank.FuelTankService = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				currentFuel -= fuelToRefill
				return nil
			},
			StatsFn: func(ctx context.Context) (*entity.FuelStat, error) {
				return &entity.FuelStat{
					FuelCapacity:    entity.FuelTankCapacity,
					FuelUsed:        entity.Fuel(10),
					LastRefuelAmout: entity.Fuel(5),
					LastRefuelAt:    time.Now(),
				}, nil
			},
		}

		fuelTank.LockService = &mock.LockService{
			LockFn:   func(ctx context.Context) {},
			NameFn:   func() string { return "lock-mock" },
			UnlockFn: func(ctx context.Context) {},
		}

		if err := fuelTank.Burn(ctx, entity.Fuel(15)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := fuelTank.Refuel(ctx, entity.Fuel(5)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stats, err := fuelTank.Stats(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if stats.FuelCapacity != entity.FuelTankCapacity {
			t.Fatalf("unexpected fuelCapacity, got: %v, want: %v", stats.FuelCapacity, entity.FuelTankCapacity)
		} else if stats.FuelUsed != 10 {
			t.Fatalf("unexpected fuelUsed, got: %v, want: %v", stats.FuelUsed, 10)
		} else if stats.LastRefuelAmout != 5 {
			t.Fatalf("unexpected lastRefuelAmout, got: %v, want: %v", stats.LastRefuelAmout, 5)
		} else if stats.LastRefuelAt.Unix() == now.Unix() {
			t.Fatalf("unexpected lastRefuelAt value, should be different from %v", now)
		}

		mgvm.SwitchToLocalFuel()

		if stats, err := fuelTank.Stats(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if stats.FuelCapacity != entity.FuelTankCapacity {
			t.Fatalf("unexpected fuelCapacity, got: %v, want: %v", stats.FuelCapacity, entity.FuelTankCapacity)
		} else if stats.FuelUsed != 10 {
			t.Fatalf("unexpected fuelUsed, got: %v, want: %v", stats.FuelUsed, 10)
		} else if stats.LastRefuelAmout != 5 {
			t.Fatalf("unexpected lastRefuelAmout, got: %v, want: %v", stats.LastRefuelAmout, 5)
		} else if stats.LastRefuelAt.Unix() == now.Unix() {
			t.Fatalf("unexpected lastRefuelAt value, should be different from %v", now)
		}
	})
}

func TestFuelTank_Concurrency(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		currentFuel := entity.Fuel(0)

		ctx := context.Background()
		mux := sync.Mutex{}

		fuelTankService := &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				currentFuel += fuel
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return currentFuel, nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				currentFuel -= fuelToRefill
				return nil
			},
		}

		lockService := &mock.LockService{
			LockFn: func(ctx context.Context) {
				mux.Lock()
			},
			NameFn: func() string {
				return "fuel-tank-mock"
			},
			UnlockFn: func(ctx context.Context) {
				mux.Unlock()
			},
		}

		tankFuelPool := []*mgvm.FuelTank{}

		for i := 0; i < 10; i++ {

			singleTankFuel := mgvm.NewFuelTank()
			singleTankFuel.FuelTankService = fuelTankService
			singleTankFuel.LockService = lockService

			tankFuelPool = append(tankFuelPool, singleTankFuel)
		}

		// this is the expected final result -> (10 * Burn of 100) - (10 * Refuel of 50)
		expectedFinalFuel := entity.Fuel(500)

		wg := sync.WaitGroup{}

		wg.Add(len(tankFuelPool))

		for _, tank := range tankFuelPool {

			go func(tb testing.TB, tank *mgvm.FuelTank) {

				tb.Helper()

				if err := tank.Burn(ctx, entity.Fuel(100)); err != nil {
					tb.Fatalf("unexpected error: %v", err)
				}

				time.Sleep(time.Millisecond * 500)

				if err := tank.Refuel(ctx, entity.Fuel(50)); err != nil {
					tb.Fatalf("unexpected error: %v", err)
				}

				wg.Done()

			}(t, tank)
		}

		wg.Wait()

		if currentFuel != expectedFinalFuel {
			t.Fatalf("unexpected currentFuel, got: %v, want: %v", currentFuel, expectedFinalFuel)
		}
	})
}

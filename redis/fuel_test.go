package redis_test

import (
	"context"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/redis"
)

func TestFuel_Burn(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		fuelTankService := redis.NewFuelTankService(db)

		fuelUsed, err := fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if err := fuelTankService.Burn(ctx, fuelUsed); err != nil {
			t.Fatal(err)
		}

		fuel, err := fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != fuelUsed {
			t.Errorf("got %d, want %d", fuel, fuelUsed)
		}

		fuelUsed = entity.Fuel(100)

		if err := fuelTankService.Burn(ctx, fuelUsed); err != nil {
			t.Fatal(err)
		}

		fuel, err = fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(100) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(100))
		}

		fuelUsed = entity.Fuel(101)

		if err := fuelTankService.Burn(ctx, fuelUsed); err != nil {
			t.Fatal(err)
		}

		fuel, err = fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(201) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(201))
		}
	})

	t.Run("ContextCancel", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx, cancel := context.WithCancel(context.Background())

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		fuelTankService := redis.NewFuelTankService(db)

		fuelUsed, err := fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		cancel()

		if err := fuelTankService.Burn(ctx, fuelUsed); err == nil {
			t.Fatal("got nil, want error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Errorf("got %v, want %v", errCode, apperr.EINTERNAL)
		}
	})
}

func TestFuel_Fuel(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		fuelTankService := redis.NewFuelTankService(db)

		fuel, err := fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(0) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(0))
		}

		if err := fuelTankService.Burn(ctx, entity.Fuel(100)); err != nil {
			t.Fatal(err)
		}

		fuel, err = fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(100) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(100))
		}
	})

	t.Run("CancelContext", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx, cancel := context.WithCancel(context.Background())

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		fuelTankService := redis.NewFuelTankService(db)

		fuel, err := fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(0) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(0))
		}

		if err := fuelTankService.Burn(ctx, entity.Fuel(100)); err != nil {
			t.Fatal(err)
		}

		cancel()

		_, err = fuelTankService.Fuel(ctx)
		if err == nil {
			t.Fatal("got nil, want error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Errorf("got %v, want %v", errCode, apperr.EINTERNAL)
		}
	})
}

func TestFuel_Refuel(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		fuelTankService := redis.NewFuelTankService(db)

		fuel, err := fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(0) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(0))
		}

		if err := fuelTankService.Burn(ctx, entity.Fuel(100)); err != nil {
			t.Fatal(err)
		}

		fuel, err = fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(100) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(100))
		}

		if err := fuelTankService.Refuel(ctx, entity.Fuel(50)); err != nil {
			t.Fatal(err)
		}

		fuel, err = fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(50) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(50))
		}

		if err := fuelTankService.Refuel(ctx, entity.Fuel(100)); err != nil {
			t.Fatal(err)
		}

		fuel, err = fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(0) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(0))
		}
	})

	t.Run("CancelContext", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx, cancel := context.WithCancel(context.Background())

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		fuelTankService := redis.NewFuelTankService(db)

		fuel, err := fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(0) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(0))
		}

		if err := fuelTankService.Burn(ctx, entity.Fuel(100)); err != nil {
			t.Fatal(err)
		}

		fuel, err = fuelTankService.Fuel(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if fuel != entity.Fuel(100) {
			t.Errorf("got %d, want %d", fuel, entity.Fuel(100))
		}

		cancel()

		if err := fuelTankService.Refuel(ctx, entity.Fuel(50)); err == nil {
			t.Fatal("got nil, want error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Errorf("got %v, want %v", errCode, apperr.EINTERNAL)
		}
	})
}

func TestFuel_Stats(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		fuelTankService := redis.NewFuelTankService(db)

		if err := fuelTankService.Burn(ctx, entity.Fuel(100)); err != nil {
			t.Fatal(err)
		}

		if err := fuelTankService.Refuel(ctx, entity.Fuel(50)); err != nil {
			t.Fatal(err)
		}

		stats, err := fuelTankService.Stats(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if stats.FuelUsed != entity.Fuel(50) {
			t.Errorf("got %d, want %d", stats.FuelUsed, entity.Fuel(50))
		}

		if stats.LastRefuelAmout != entity.Fuel(50) {
			t.Errorf("got %d, want %d", stats.LastRefuelAmout, entity.Fuel(50))
		}
	})

	t.Run("InitialStateStats", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx := context.Background()

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		fuelTankService := redis.NewFuelTankService(db)

		stats, err := fuelTankService.Stats(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if stats.FuelUsed != entity.Fuel(0) {
			t.Errorf("got %d, want %d", stats.FuelUsed, entity.Fuel(0))
		}

		if stats.LastRefuelAmout != entity.Fuel(0) {
			t.Errorf("got %d, want %d", stats.LastRefuelAmout, entity.Fuel(0))
		}
	})

	t.Run("CancelContext", func(t *testing.T) {

		db := MustOpenDB(t)
		defer db.Close()

		ctx, cancel := context.WithCancel(context.Background())

		if err := db.FlushAll(ctx); err != nil {
			t.Fatal(err)
		}

		fuelTankService := redis.NewFuelTankService(db)

		cancel()

		if _, err := fuelTankService.Stats(ctx); err == nil {
			t.Fatal("got nil, want error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Errorf("got %v, want %v", errCode, apperr.EINTERNAL)
		}
	})
}

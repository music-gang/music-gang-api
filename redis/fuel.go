package redis

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.FuelTankService = (*FuelTankService)(nil)

// Defines all keys used by redis.
const (
	fuelUsedRedisKey = "mgvm_fuel_tank_fuel_used"
	fuelCapRedisKey  = "mgvm_fuel_tank_fuel_cap"
)

// FuelTankService implements the FuelTankService interface.
// It is used to manage the fuel tank shared data.
type FuelTankService struct {
	db *DB
}

// NewFueltankService creates a new FuelTankService.
func NewFueltankService(db *DB) *FuelTankService {
	return &FuelTankService{db: db}
}

// Burn consumes the specified amount of fuel.
func (ft *FuelTankService) Burn(ctx context.Context, fuelUsed entity.Fuel) error {
	return burn(ctx, ft.db, fuelUsed)
}

// Cap returns the max capacity of the fuel tank.
func (ft *FuelTankService) Cap(ctx context.Context) (entity.Fuel, error) {
	return cap(ctx, ft.db)
}

// Fuel returns the current amount of fuel used.
func (ft *FuelTankService) Fuel(ctx context.Context) (entity.Fuel, error) {
	return fuel(ctx, ft.db)
}

// Refuel refills the fuel tank by the specified amount.
func (ft *FuelTankService) Refuel(ctx context.Context, fuelToRefill entity.Fuel) error {
	return refuel(ctx, ft.db, fuelToRefill)
}

// cap returns the max capacity of the fuel tank.
// It is not thread-safe, use it with some lock service.
func cap(ctx context.Context, db *DB) (entity.Fuel, error) {

	rawVal, err := db.client.Get(ctx, fuelCapRedisKey).Result()
	if err == redis.Nil {
		return entity.Fuel(0), nil
	} else if err != nil {
		return entity.Fuel(0), apperr.Errorf(apperr.EINTERNAL, "failed to get fuel cap from redis: %v", err)
	}

	val, err := strconv.ParseUint(rawVal, 10, 64)
	if err != nil {
		return entity.Fuel(0), apperr.Errorf(apperr.EINTERNAL, "failed to parse fuel cap from redis: %v", err)
	}

	return entity.Fuel(val), nil
}

// burn consumes the specified amount of fuel.
// It is not thread-safe, use it with some lock service.
func burn(ctx context.Context, db *DB, fuelUsed entity.Fuel) error {

	currentFuelUsed, err := fuel(ctx, db)
	if err != nil {
		return err
	}

	// add the fuel used to the current fuel used
	newFuelUsed := currentFuelUsed + fuelUsed

	if err := db.client.Set(ctx, fuelUsedRedisKey, newFuelUsed, 0).Err(); err != nil {
		return err
	}

	return nil
}

// fuel returns the current amount of fuel used.
// It is not thread-safe, use it with some lock service.
func fuel(ctx context.Context, db *DB) (entity.Fuel, error) {

	rawVal, err := db.client.Get(ctx, fuelUsedRedisKey).Result()
	if err == redis.Nil {
		// The key does not exist.
		return entity.Fuel(0), nil
	} else if err != nil {
		return entity.Fuel(0), apperr.Errorf(apperr.EINTERNAL, "failed to get fuel used from redis: %v", err)
	}

	val, err := strconv.ParseUint(rawVal, 10, 64)
	if err != nil {
		return entity.Fuel(0), apperr.Errorf(apperr.EINTERNAL, "failed to parse fuel used from redis: %v", err)
	}

	return entity.Fuel(val), nil
}

// refuel refills the fuel tank by the specified amount.
// It is not thread-safe, use it with some lock service.
func refuel(ctx context.Context, db *DB, fuelToRefill entity.Fuel) error {

	currentFuelUsed, err := fuel(ctx, db)
	if err != nil {
		return err
	}

	// check if the fuel to refill is greater than the current fuel used, if so, set refuel to current fuel used
	if fuelToRefill > currentFuelUsed {
		fuelToRefill = currentFuelUsed
	}

	// remove the fuel refilled from the current fuel used

	newFuelUsed := currentFuelUsed - fuelToRefill

	if err := db.client.Set(ctx, fuelUsedRedisKey, newFuelUsed, 0).Err(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to refill fuel tank in redis: %v", err)
	}

	return nil
}

package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.FuelTankService = (*FuelTankService)(nil)

// Defines all keys used by redis.
const (
	FuelUsedRedisKey     = "mgvm_fuel_tank_fuel_used"
	FuelLastRefillAmount = "mgvm_fuel_tank_fuel_last_refill_amount"
	FuelLastRefillTime   = "mgvm_fuel_tank_fuel_last_refill_time"
)

// FuelTankService implements the FuelTankService interface.
// It is used to manage the fuel tank shared data.
type FuelTankService struct {
	db *DB
}

// NewFuelTankService creates a new FuelTankService.
func NewFuelTankService(db *DB) *FuelTankService {
	return &FuelTankService{db: db}
}

// Burn consumes the specified amount of fuel.
func (ft *FuelTankService) Burn(ctx context.Context, fuelUsed entity.Fuel) error {
	return burn(ctx, ft.db, fuelUsed)
}

// Fuel returns the current amount of fuel used.
func (ft *FuelTankService) Fuel(ctx context.Context) (entity.Fuel, error) {
	return fuel(ctx, ft.db)
}

// Refuel refills the fuel tank by the specified amount.
func (ft *FuelTankService) Refuel(ctx context.Context, fuelToRefill entity.Fuel) error {
	return refuel(ctx, ft.db, fuelToRefill)
}

// Stats returns the current fuel tank stats.
func (ft *FuelTankService) Stats(ctx context.Context) (*entity.FuelStat, error) {
	return stats(ctx, ft)
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

	if err := db.client.Set(ctx, FuelUsedRedisKey, newFuelUsed, 0).Err(); err != nil {
		return err
	}

	return nil
}

// fuel returns the current amount of fuel used.
// It is not thread-safe, use it with some lock service.
func fuel(ctx context.Context, db *DB) (entity.Fuel, error) {

	rawVal, err := db.client.Get(ctx, FuelUsedRedisKey).Result()
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

	if err := db.client.Set(ctx, FuelUsedRedisKey, newFuelUsed, 0).Err(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to refill fuel tank in redis: %v", err)
	}
	if err := db.client.Set(ctx, FuelLastRefillAmount, fuelToRefill, 0).Err(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to set last fuel refill amount in redis: %v", err)
	}
	if err := db.client.Set(ctx, FuelLastRefillTime, time.Now().Unix(), 0).Err(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to set last fuel refill time in redis: %v", err)
	}

	return nil
}

// stats returns the current fuel tank stats.
// It is not thread-safe, use it with some lock service.
func stats(ctx context.Context, ft *FuelTankService) (*entity.FuelStat, error) {

	var lastRefillAmount entity.Fuel
	var lastRefillTime time.Time

	fuel, err := ft.Fuel(ctx)
	if err != nil {
		return nil, err
	}

	rawLastRefillAmount, err := ft.db.client.Get(ctx, FuelLastRefillAmount).Result()
	if err == redis.Nil {
		lastRefillAmount = entity.Fuel(0)
	} else if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to get last refill amount from redis: %v", err)
	} else {
		val, err := strconv.ParseUint(rawLastRefillAmount, 10, 64)
		if err != nil {
			return nil, apperr.Errorf(apperr.EINTERNAL, "failed to parse last refill amount from redis: %v", err)
		}

		lastRefillAmount = entity.Fuel(val)
	}

	rawLastRefillTime, err := ft.db.client.Get(ctx, FuelLastRefillTime).Result()
	if err == redis.Nil {
		lastRefillTime = time.Time{}
	} else if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to get last refill time from redis: %v", err)
	} else {
		val, err := strconv.ParseInt(rawLastRefillTime, 10, 64)
		if err != nil {
			return nil, apperr.Errorf(apperr.EINTERNAL, "failed to parse last refill time from redis: %v", err)
		}

		lastRefillTime = time.Unix(val, 0).UTC()
	}

	return &entity.FuelStat{
		FuelCapacity:    entity.FuelTankCapacity,
		FuelUsed:        fuel,
		LastRefuelAmout: lastRefillAmount,
		LastRefuelAt:    lastRefillTime,
	}, nil
}

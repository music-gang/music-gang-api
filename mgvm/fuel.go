package mgvm

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.FuelTankService = (*FuelTank)(nil)

// useRemoteFuel is a flag that indicates if the fuel tank should be synchronized with the remote service.
// TODO: this should be dynamic, maybe based on context param.
var useRemoteFuel = true

// FuelTank is the fuel tank of the MusicGang VM.
//
// His purpose in to manage the fuel usage by the VM.
// Implements the FuelTankService interface for local counter and delegate the synchronized counter to internal.
// It is thread-safe.
//
// The goal is to have a local counter and a remote counter.
// The local counter is used when the sync is not required, maybe in a read-only operation.
// The remote counter is used when the sync is required, maybe in a write operation to keep the consistency between various instances.
// The goal is to have this virtual fuel tank where one property is his capacity and the other is his fuel used.
// The capacity rappresent the max fuel used by the virtual fuel tank.
// If the capacity is reached, the virtual fuel tank is not able to consume more fuel, in this case the virtual fuel tank needs to be refilled.
// The fuel used rappresent the current fuel used by the virtual fuel tank and it should be in the range [0, capacity].
// Is legal to have the fuel used greater than the capacity, but in this case the MusicGang VM will stop until the fuel tank is refilled.
// In this cases the virtual fuel tank is not able to consume more fuel (capacity reached or more).
//
// To refill the virtual fuel tank is required to call the Refuel method.
// FuelTank is not able to refill itself automatically, this workload is delegated to MusicGang VM.
type FuelTank struct {
	// FuelTankService is the service for managing the fuel tank.
	// FuelTank delegates to this service to achieve scalability.
	FuelTankService service.FuelTankService
	LockService     service.LockService

	// lastRefuelAmount is the last amount of fuel used to refill the virtual fuel tank.
	// It should be used only in case the sync is not required.
	lastRefuelAmount entity.Fuel
	// LastRefuelAt is the last time the virtual fuel tank was refilled.
	// It should be used only in case the sync is not required.
	LastRefuelAt int64
	// localFuelUsed is the amount of fuel used locally.
	// It should used only in case the sync is not required.
	localFuelUsed entity.Fuel
}

// NewFuelTank creates a new FuelTank.
func NewFuelTank() *FuelTank {
	return &FuelTank{}
}

// Burn consumes the specified amount of fuel.
func (ft *FuelTank) Burn(ctx context.Context, fuel entity.Fuel) error {
	return burn(ctx, ft, fuel)
}

// Fuel returns the current amount of fuel used.
func (ft *FuelTank) Fuel(ctx context.Context) (entity.Fuel, error) {
	return fuel(ctx, ft, !useRemoteFuel)
}

// Refuel refills the virtual fuel tank with the specified amount of fuel.
func (ft *FuelTank) Refuel(ctx context.Context, fuelToRefill entity.Fuel) error {
	return refuel(ctx, ft, fuelToRefill)
}

// Stats returns the current amount of fuel used.
func (ft *FuelTank) Stats(ctx context.Context) (*entity.FuelStat, error) {
	return stats(ctx, ft, useRemoteFuel)
}

// localFuel returns the current amount of fuel used from the local counter.
// It should used only in case the sync is not required.
// It is thread-safe.
func (ft *FuelTank) localFuel() entity.Fuel {
	return entity.Fuel(atomic.LoadUint64((*uint64)(&ft.localFuelUsed)))
}

// burn consumes the specified amount of fuel.
// It sync the fuel tank between the local counter and the remote service.
func burn(ctx context.Context, ft *FuelTank, fuel entity.Fuel) error {

	// first, we need to aquire the lock

	ft.LockService.Lock(ctx)
	defer ft.LockService.Unlock(ctx)

	// second, we need to retrive the current fuel tank capacity and the current fuel used

	fuelUsed, err := ft.Fuel(ctx)
	if err != nil {
		return err
	}

	// third, we need to check if the current fuel used + the passed fuel is greater than the max fuel tank capacity
	if fuelUsed+fuel > entity.FuelTankCapacity {
		return service.ErrFuelTankNotEnough
	}

	// fourth, we need to update the synchronized fuel tank
	if err := ft.FuelTankService.Burn(ctx, fuel); err != nil {
		return err
	}

	// five, we need to update the local fuel tank and capacity
	atomic.AddUint64((*uint64)(&ft.localFuelUsed), uint64(fuel))

	return nil
}

// fuel returns the current amount of fuel used.
// If local is true, it returns the current fuel used from the local counter, otherwise it returns the current fuel used from the remote service.
func fuel(ctx context.Context, ft *FuelTank, local bool) (entity.Fuel, error) {
	if local {
		return ft.localFuel(), nil
	}
	return ft.FuelTankService.Fuel(ctx)
}

// refuel refills the fuel tank by the specified amount.
// It sync the fuel tank between the local counter and the remote service.
// If the passed fuel to refill is greater then actual fuel used, it sets fuel used to 0.
func refuel(ctx context.Context, ft *FuelTank, refillFuel entity.Fuel) error {

	// first, we need to aquire the lock

	ft.LockService.Lock(ctx)
	defer ft.LockService.Unlock(ctx)

	// second, we need to retrive the current fuel tank capacity and the current fuel used

	fuelUsed, err := ft.Fuel(ctx)
	if err != nil {
		return err
	}

	// third, we need to check if the fuel to refill is not greater than fuel used, otherwise 0 is set

	if refillFuel > fuelUsed {
		refillFuel = fuelUsed
	}

	fuelUsedAfterRefill := fuelUsed - refillFuel

	// fourth, we need to update the synchronized fuel tank
	if err := ft.FuelTankService.Refuel(ctx, refillFuel); err != nil {
		return err
	}

	// fifth, we need to update the local fuel tank and capacity
	atomic.StoreUint64((*uint64)(&ft.localFuelUsed), uint64(fuelUsedAfterRefill))
	atomic.StoreUint64((*uint64)(&ft.lastRefuelAmount), uint64(refillFuel))
	atomic.StoreInt64((*int64)(&ft.LastRefuelAt), time.Now().Unix())

	return nil
}

// Stats returns the current amount of fuel used.
func stats(ctx context.Context, ft *FuelTank, useRemoteFuel bool) (*entity.FuelStat, error) {

	if useRemoteFuel {
		ft.LockService.Lock(ctx)
		defer ft.LockService.Unlock(ctx)
		return ft.FuelTankService.Stats(ctx)
	}

	fuel := ft.localFuel()

	lastRefuelAmout := entity.Fuel(atomic.LoadUint64((*uint64)(&ft.lastRefuelAmount)))
	lastRefuelAt := atomic.LoadInt64(&ft.LastRefuelAt)

	stats := &entity.FuelStat{
		FuelCapacity:    entity.FuelTankCapacity,
		FuelUsed:        fuel,
		LastRefuelAmout: lastRefuelAmout,
		LastRefuelAt:    time.Unix(lastRefuelAt, 0),
	}

	return stats, nil
}

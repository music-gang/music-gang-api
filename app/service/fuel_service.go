package service

import (
	"context"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
)

// ErrFuelTankCapacity is returned when the initial capacity is greater than the max capacity.
var ErrFuelTankNotEnough = apperr.Errorf(apperr.EINTERNAL, "fuel tank is not enough")

// FuelTanker is the interface for the fuel tank.
type FuelTankService interface {
	// Burn consumes the specified amount of fuel.
	Burn(ctx context.Context, fuel entity.Fuel) error
	// Cap returns the max capacity of the fuel tank.
	Cap(ctx context.Context) (entity.Fuel, error)
	// Fuel returns the current amount of fuel used.
	Fuel(ctx context.Context) (entity.Fuel, error)
	// Refuel refills the fuel tank by the specified amount.
	Refuel(ctx context.Context, fuelToRefill entity.Fuel) error
}

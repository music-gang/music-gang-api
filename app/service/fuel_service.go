package service

import (
	"context"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
)

// ErrFuelTankCapacity is returned when the initial capacity is greater than the max capacity.
var ErrFuelTankNotEnough = apperr.Errorf(apperr.EMGVM_LOWFUEL, "fuel tank is not enough")

type FuelService interface {
	// Fuel returns the current amount of fuel used.
	Fuel(ctx context.Context) (entity.Fuel, error)
}

// FuelTanker is the interface for the fuel tank.
type FuelTankService interface {
	FuelService

	// FuelTankService can implements the FuelStatsService
	FuelStatsService

	// Burn consumes the specified amount of fuel.
	Burn(ctx context.Context, fuel entity.Fuel) error
	// Refuel refills the fuel tank by the specified amount.
	Refuel(ctx context.Context, fuelToRefill entity.Fuel) error
}

// FuelStationService is the interface for the fuel station.
type FuelStationService interface {
	// IsRunning returns true if the FuelStation is running
	IsRunning() bool
	// ResumeRefueling starts the FuelStation.
	// It will start refueling the fuel tank every FuelRefillRate.
	// If the FuelStation is already running, it will return an error.
	ResumeRefueling(ctx context.Context) error
	// StopRefueling stops the FuelStation.
	// If the FuelStation is not running, it will return an error.
	StopRefueling(ctx context.Context) error
}

// FuelStatsService is the interface for the fuel meter.
// FuelStatsService returns the current amount of fuel used.
type FuelStatsService interface {
	// Stats returns the current amount of fuel used.
	Stats(ctx context.Context) (*entity.FuelStat, error)
}

// FuelMonitorService is the interface for the fuel monitor.
// This service monitors the fuel consumption of the engine and should send events when reach safe levels or when the fuel usage is too high.
type FuelMonitorService interface {
	// StartMonitoring starts the fuel monitor.
	StartMonitoring(ctx context.Context) error
	// StopMonitoring stops the fuel monitor.
	StopMonitoring(ctx context.Context) error
}

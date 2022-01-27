package service

import "context"

const (
	FuelTankLockName = "fuel_tank_mux"
)

// LockService is the interface for the lock service.
// It is used to implement a distributed lock.
type LockService interface {
	// Lock acquires the lock.
	Lock(ctx context.Context)
	// Name returns the name of the lock.
	Name() string
	// Unlock releases the lock.
	Unlock(ctx context.Context)
}

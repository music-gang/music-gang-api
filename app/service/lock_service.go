package service

import "context"

const (
	FuelTankLockName = "fuel_tank_mux"
)

// LockService is the interface for the lock service.
// It is used to implement a distributed lock.
type LockService interface {
	// LockContext acquires the lock.
	LockContext(ctx context.Context) error
	// Name returns the name of the lock.
	Name() string
	// UnlockContext releases the lock.
	UnlockContext(ctx context.Context) (bool, error)
}

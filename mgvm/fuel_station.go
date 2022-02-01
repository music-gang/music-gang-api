package mgvm

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

// FuelStation is a fuel station that can be used to refuel the fuel tank.
// FuelStation is responsible for starting and stopping the refueling of the fuel tank.
type FuelStation struct {
	FuelTankService service.FuelTankService
	LogService      service.LogService

	FuelRefillAmount entity.Fuel
	FuelRefillRate   time.Duration

	running int32
}

// NewFuelStation creates a new FuelStation
func NewFuelStation() *FuelStation {
	return &FuelStation{}
}

// IsRunning returns true if the FuelStation is running
func (fs *FuelStation) IsRunning() bool {
	return atomic.LoadInt32(&fs.running) == 1
}

// ResumeRefueling starts the FuelStation.
// It will start refueling the fuel tank every FuelRefillRate.
// If the FuelStation is already running, it will return an error.
func (fs *FuelStation) ResumeRefueling(ctx context.Context) error {
	if fs.IsRunning() {
		return apperr.Errorf(apperr.EMGVM, "FuelStation is already running")
	}
	go resumeRefueling(ctx, fs)
	return nil
}

// StopRefueling stops the FuelStation.
// If the FuelStation is not running, it will return an error.
func (fs *FuelStation) StopRefueling(ctx context.Context) error {
	if !fs.IsRunning() {
		return apperr.Errorf(apperr.EMGVM, "FuelStation is not running")
	}
	return stopRefueling(ctx, fs)
}

// setRunningState sets the running state of the FuelStation.
func (fs *FuelStation) setRunningState(val int32) {
	atomic.StoreInt32(&fs.running, val)
}

// resumeRefueling starts the FuelStation.
func resumeRefueling(ctx context.Context, fs *FuelStation) error {

	fs.setRunningState(1)

	ticker := time.NewTicker(fs.FuelRefillRate)
	defer ticker.Stop()

	for {

		if !fs.IsRunning() {
			// the FuelStation is stopped
			return nil
		}

		select {
		case <-ctx.Done():
			fs.setRunningState(0)
			return nil
		case <-ticker.C:
			if err := internalRefueler(ctx, fs); err != nil {
				fs.LogService.ReportError(ctx, err)
			}
		}
	}
}

// stopRefueling stops the FuelStation.
func stopRefueling(ctx context.Context, fs *FuelStation) error {
	fs.setRunningState(0)
	return nil
}

// internalRefueler refuels the fuel tank.
func internalRefueler(ctx context.Context, fs *FuelStation) error {

	defer func() {
		if r := recover(); r != nil {
			fs.LogService.ReportPanic(ctx, r)
			return
		}
	}()

	if err := fs.FuelTankService.Refuel(ctx, fs.FuelRefillAmount); err != nil {
		return err
	}

	return nil
}

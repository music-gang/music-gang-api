package mgvm

import (
	"context"
	"time"

	log "github.com/inconshreveable/log15"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/common"
)

var _ service.FuelStationService = (*FuelStation)(nil)

// FuelStation is a fuel station that can be used to refuel the fuel tank.
// FuelStation is responsible for starting and stopping the refueling of the fuel tank.
type FuelStation struct {
	common.RunningState

	FuelTankService service.FuelTankService
	LogService      log.Logger

	FuelRefillAmount entity.Fuel
	FuelRefillRate   time.Duration
}

// NewFuelStation creates a new FuelStation
func NewFuelStation() *FuelStation {
	return &FuelStation{}
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

// resumeRefueling starts the FuelStation.
func resumeRefueling(ctx context.Context, fs *FuelStation) error {

	fs.SetRunningState(1)

	ticker := time.NewTicker(fs.FuelRefillRate)
	defer ticker.Stop()

	for {

		if !fs.IsRunning() {
			// the FuelStation is stopped
			return nil
		}

		select {
		case <-ctx.Done():
			fs.SetRunningState(0)
			return nil
		case <-ticker.C:
			if err := internalRefueler(ctx, fs); err != nil {
				fs.LogService.Error(apperr.ErrorLog(err))
			}
		}
	}
}

// stopRefueling stops the FuelStation.
func stopRefueling(ctx context.Context, fs *FuelStation) error {
	fs.SetRunningState(0)
	return nil
}

// internalRefueler refuels the fuel tank.
func internalRefueler(ctx context.Context, fs *FuelStation) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = apperr.Errorf(apperr.EMGVM, "internal refueler panic: %v", r)
		}
	}()

	if err := fs.FuelTankService.Refuel(ctx, fs.FuelRefillAmount); err != nil {
		return err
	}

	return nil
}

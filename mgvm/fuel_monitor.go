package mgvm

import (
	"context"
	"time"

	log "github.com/inconshreveable/log15"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/app/util"
	"github.com/music-gang/music-gang-api/event"
)

var _ service.FuelMonitorService = (*FuelMonitor)(nil)

// FuelMonitor is the fuel monitor.
// It measures the fuel level of the Music Gang Virtual Machine.
// When the fuel reach a specific threshold, an event is published and this event should be handled by the VM.
// When the fuel reach a safe level, an event is published and this event should be handled by the VM.
type FuelMonitor struct {
	util.RunningState

	EventService       *event.EventService
	EngineStateService service.EngineStateService
	FuelService        service.FuelService

	LogService log.Logger
}

// NewFuelMonitor creates a new FuelMonitor.
func NewFuelMonitor() *FuelMonitor {
	return &FuelMonitor{}
}

// StartMonitoring implements service.FuelMonitorService
func (fm *FuelMonitor) StartMonitoring(ctx context.Context) error {
	if fm.IsRunning() {
		return apperr.Errorf(apperr.EMGVM, "FuelMonitor is already running")
	}
	go startMonitoring(ctx, fm)
	return nil
}

// StopMonitoring implements service.FuelMonitorService
func (fm *FuelMonitor) StopMonitoring(ctx context.Context) error {
	if !fm.IsRunning() {
		return apperr.Errorf(apperr.EMGVM, "FuelMonitor is not running")
	}
	return stopMonitoring(ctx, fm)
}

// startMonitoring starts the fuel monitoring.
func startMonitoring(ctx context.Context, fm *FuelMonitor) error {

	fm.SetRunningState(1)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {

		if !fm.IsRunning() {
			return nil
		}

		select {

		case <-ctx.Done():
			fm.SetRunningState(0)
			return nil
		case <-ticker.C:
			if err := meter(ctx, fm); err != nil {
				fm.LogService.Error(apperr.ErrorLog(err))
				return err
			}
		}
	}
}

// stopMonitoring stops the fuel monitoring.
func stopMonitoring(ctx context.Context, fm *FuelMonitor) error {
	fm.SetRunningState(0)
	return nil
}

// meter measures the fuel level of the engine.
// When the fuel level reach a threshold of 95%, an event is published.
// When the fuel level reach a safe level, an event is published.
func meter(ctx context.Context, fm *FuelMonitor) error {

	defer func() {
		if r := recover(); r != nil {
			fm.LogService.Crit("Panic while measuring fuel consumption", log.Ctx{"panic": r})
		}
	}()

	if fm.EngineStateService.State() == entity.StatePaused {
		if fuel, err := fm.FuelService.Fuel(ctx); err != nil {
			return err
		} else if float64(fuel) <= float64(entity.FuelTankCapacity)*0.65 {

			fm.EventService.PublishEvent(ctx, event.Event{
				Type:    event.EngineShouldResumeEvent,
				Message: "Resume engine due to reaching safe fuel level",
			})

		}
	} else if fm.EngineStateService.State() == entity.StateRunning {
		if fuel, err := fm.FuelService.Fuel(ctx); err != nil {
			return err
		} else if float64(fuel) >= float64(entity.FuelTankCapacity)*0.95 {

			fm.EventService.PublishEvent(ctx, event.Event{
				Type:    event.EngineShouldPauseEvent,
				Message: "Pause engine due to excessive fuel consumption",
			})
		}
	}

	return nil
}

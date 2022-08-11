package mgvm_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/event"
	"github.com/music-gang/music-gang-api/mgvm"
	"github.com/music-gang/music-gang-api/mock"
)

func TestFuelMonitor_StartMonitor(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		engineState := entity.StateRunning
		currentFuel := entity.Fuel(0)

		defer cancel()

		fuelMonitor := mgvm.NewFuelMonitor()

		fuelMonitor.EngineStateService = &mock.EngineService{
			StateFn: func() entity.VmState {
				return entity.VmState(atomic.LoadInt32((*int32)(&engineState)))
			},
		}
		fuelMonitor.LogService = &mock.LoggerNoOp{}
		fuelMonitor.FuelService = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return entity.Fuel(atomic.LoadUint64((*uint64)(&currentFuel))), nil
			},
		}
		fuelMonitor.EventService = event.NewEventService()

		subToResumeEngine := fuelMonitor.EventService.Subscribe(ctx, event.EngineShouldResumeEvent)
		subToPauseEngine := fuelMonitor.EventService.Subscribe(ctx, event.EngineShouldPauseEvent)

		receivedResumeEvent := atomic.Bool{}
		receivedPauseEvent := atomic.Bool{}

		go func() {

			for {
				select {
				case <-ctx.Done():
					return
				case <-subToResumeEngine.C():
					receivedResumeEvent.Store(true)
				case <-subToPauseEngine.C():
					receivedPauseEvent.Store(true)
				}
			}
		}()

		if err := fuelMonitor.StartMonitoring(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		atomic.StoreUint64((*uint64)(&currentFuel), uint64(float64(entity.FuelTankCapacity)*0.96))

		time.Sleep(1 * time.Second)

		atomic.StoreUint64((*uint64)(&currentFuel), uint64(0))
		atomic.StoreInt32((*int32)(&engineState), int32(entity.StatePaused))

		time.Sleep(1 * time.Second)

		if !receivedResumeEvent.Load() || !receivedPauseEvent.Load() {
			t.Errorf("Expected resume and pause events to be received, got %v and %v", receivedResumeEvent, receivedPauseEvent)
		}

		fuelMonitor.SetRunningState(0)

		time.Sleep(1 * time.Second)

		if fuelMonitor.IsRunning() {
			t.Errorf("Expected fuel monitor to be stopped, got %v", fuelMonitor.IsRunning())
		}
	})

	t.Run("ContextCanceled", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		fuelMonitor := mgvm.NewFuelMonitor()

		cancel()

		if err := fuelMonitor.StartMonitoring(ctx); err != nil {
			// Any way the fuel monitor should starts, but stop immediately in the goroutine if the context is canceled
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(1 * time.Second)

		if fuelMonitor.IsRunning() {
			t.Errorf("Expected fuel monitor to be stopped, got %v", fuelMonitor.IsRunning())
		}
	})

	t.Run("AlreadyRunning", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel()

		fuelMonitor := mgvm.NewFuelMonitor()

		fuelMonitor.SetRunningState(1)

		if err := fuelMonitor.StartMonitoring(ctx); err == nil {
			t.Errorf("Expected error, got nil")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Expected error code %v, got %v", apperr.EMGVM, errCode)
		}
	})

	t.Run("FuelErrOnEngineOnEnginePause", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		engineState := entity.StatePaused
		errorOccurred := atomic.Bool{}

		defer cancel()

		fuelMonitor := mgvm.NewFuelMonitor()

		fuelMonitor.EngineStateService = &mock.EngineService{
			StateFn: func() entity.VmState {
				return engineState
			},
		}
		fuelMonitor.LogService = &mock.LoggerNoOp{
			ErrorFn: func(msg string, ctx ...interface{}) {
				errorOccurred.Store(true)
			},
		}
		fuelMonitor.FuelService = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return 0, apperr.Errorf(apperr.EMGVM, "Fuel error")
			},
		}
		fuelMonitor.EventService = event.NewEventService()

		if err := fuelMonitor.StartMonitoring(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(1 * time.Second)

		if !errorOccurred.Load() {
			t.Errorf("Expected error to be logged, got none")
		}
	})

	t.Run("FuelErrOnEngineOnEngineRunning", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		engineState := entity.StateRunning
		errorOccurred := atomic.Bool{}

		defer cancel()

		fuelMonitor := mgvm.NewFuelMonitor()

		fuelMonitor.EngineStateService = &mock.EngineService{
			StateFn: func() entity.VmState {
				return engineState
			},
		}
		fuelMonitor.LogService = &mock.LoggerNoOp{
			ErrorFn: func(msg string, ctx ...interface{}) {
				errorOccurred.Store(true)
			},
		}
		fuelMonitor.FuelService = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return 0, apperr.Errorf(apperr.EMGVM, "Fuel error")
			},
		}
		fuelMonitor.EventService = event.NewEventService()

		if err := fuelMonitor.StartMonitoring(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(1 * time.Second)

		if !errorOccurred.Load() {
			t.Errorf("Expected error to be logged, got none")
		}
	})

	t.Run("Panic", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		engineState := entity.StateRunning
		panicOccurred := atomic.Bool{}

		defer cancel()

		fuelMonitor := mgvm.NewFuelMonitor()

		fuelMonitor.EngineStateService = &mock.EngineService{
			StateFn: func() entity.VmState {
				return engineState
			},
		}
		fuelMonitor.EventService = event.NewEventService()
		fuelMonitor.LogService = &mock.LoggerNoOp{
			CritFn: func(msg string, ctx ...interface{}) {
				panicOccurred.Store(true)
			},
		}

		if err := fuelMonitor.StartMonitoring(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(1 * time.Second)

		if !panicOccurred.Load() {
			t.Errorf("Expected panic to be logged, got none")
		}
	})
}

func TestFuelMonitor_StopMonitoring(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		ctx := context.Background()

		fuelMonitor := mgvm.NewFuelMonitor()

		fuelMonitor.SetRunningState(1)

		if err := fuelMonitor.StopMonitoring(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})

	t.Run("AlreadyStopped", func(t *testing.T) {

		ctx := context.Background()

		fuelMonitor := mgvm.NewFuelMonitor()

		if err := fuelMonitor.StopMonitoring(ctx); err == nil {
			t.Errorf("Expected error, got nil")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Expected error code %v, got %v", apperr.EMGVM, errCode)
		}
	})
}

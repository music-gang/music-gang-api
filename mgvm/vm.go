package mgvm

import (
	"context"
	"sync"
	"time"

	log "github.com/inconshreveable/log15"
	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/event"
)

var _ service.VmService = (*MusicGangVM)(nil)
var _ service.VmCallableService = (*MusicGangVM)(nil)

// VmFunc is a generic function callback executed by the vm.
type VmFunc func(ctx context.Context, ref service.VmCallable) (interface{}, error)

// MusicGangVM is a virtual machine for the Mg language(nodeJS for now).
type MusicGangVM struct {
	ctx    context.Context
	cancel context.CancelFunc

	engineShouldResumeSub *event.Subscription
	engineShouldPauseSub  *event.Subscription

	*sync.Cond

	LogService log.Logger

	EventService *event.EventService

	EngineService   service.EngineService
	FuelTank        service.FuelTankService
	FuelStation     service.FuelStationService
	FuelMonitor     service.FuelMonitorService
	CPUsPoolService service.CPUsPoolService

	AuthManagmentService     service.AuthManagmentService
	ContractManagmentService service.ContractManagmentService
	UserManagmentService     service.UserManagmentService
	StateService             service.StateService
	CacheStateService        service.StateCacheService
}

// MusicGangVM creates a new MusicGangVM.
// It should be called only once.
func NewMusicGangVM() *MusicGangVM {
	ctx := app.NewContextWithTags(context.Background(), []string{app.ContextTagMGVM})
	ctx, cancel := context.WithCancel(ctx)
	return &MusicGangVM{
		ctx:    ctx,
		cancel: cancel,
		Cond:   sync.NewCond(&sync.Mutex{}),
		// Create a default event service, later we can replace it with a real one.
		EventService: event.NewEventService(),
	}
}

// Run starts the vm.
func (vm *MusicGangVM) Run() (err error) {

	vm.engineShouldResumeSub = vm.EventService.Subscribe(vm.ctx, event.EngineShouldResumeEvent)
	vm.engineShouldPauseSub = vm.EventService.Subscribe(vm.ctx, event.EngineShouldPauseEvent)

	if err := vm.FuelStation.ResumeRefueling(vm.ctx); err != nil {
		return err
	}

	if err := vm.FuelMonitor.StartMonitoring(vm.ctx); err != nil {
		return err
	}

	if err := vm.Resume(); err != nil {
		return err
	}

	go func() {

		for {
			select {
			case <-vm.ctx.Done():
				return
			case e := <-vm.engineShouldResumeSub.C():
				vm.LogService.Info(e.Message)
				vm.Resume()
			case e := <-vm.engineShouldPauseSub.C():
				vm.LogService.Info(e.Message)
				vm.Pause()
			}
		}
	}()

	return nil
}

// Close closes the vm.
func (vm *MusicGangVM) Close() error {

	if vm.State() == entity.StateStopped {
		return apperr.Errorf(apperr.EMGVM, "VM is already closed")
	}

	if vm.State() == entity.StateInitializing {
		return apperr.Errorf(apperr.EMGVM, "VM is still initializing")
	}

	vm.cancel()

	vm.engineShouldPauseSub.Close()
	vm.engineShouldResumeSub.Close()

	if err := vm.EngineService.Stop(); err != nil {
		return err
	}

	if err := vm.FuelStation.StopRefueling(vm.ctx); err != nil {
		return err
	}

	if err := vm.FuelMonitor.StopMonitoring(vm.ctx); err != nil {
		return err
	}

	return nil
}

// IsRunning returns true if the engine is running.
// Delegates to the engine service.
func (vm *MusicGangVM) IsRunning() bool {
	return vm.EngineService.IsRunning()
}

// Pause pauses the engine.
// Delegates to the engine service.
func (vm *MusicGangVM) Pause() error {
	return vm.EngineService.Pause()
}

// Resume resumes the engine.
// Delegates to the engine service.
func (vm *MusicGangVM) Resume() error {
	if err := vm.EngineService.Resume(); err != nil {
		return err
	}
	vm.Broadcast()
	return nil
}

// State returns the state of the engine.
// Delegates to the engine service.
func (vm *MusicGangVM) State() entity.VmState {
	return vm.EngineService.State()
}

// Stop stops the engine.
// Delegates to the engine service.
func (vm *MusicGangVM) Stop() error {
	return vm.EngineService.Stop()
}

// makeOperation executes the given operations.
func (vm *MusicGangVM) makeOperation(ctx context.Context, ref service.VmCallable, fn VmFunc) (res interface{}, err error) {
	select {
	case <-ctx.Done():
		return nil, apperr.Errorf(apperr.EMGVM, "Timeout while executing operation")
	default:
		if ref.WithEngineState() {
			func() {
				vm.L.Lock()
				for !vm.IsRunning() {
					vm.LogService.Info("Wait for engine to resume")
					vm.Wait()
				}
				vm.L.Unlock()
			}()
		}
	}

	defer func() {
		// handle engine timeout or panic
		if r := recover(); r != nil {
			if r == service.EngineExecutionTimeoutPanic {
				err = apperr.Errorf(apperr.EMGVM, "Timeout while executing operation")
				return
			}
			err = apperr.Errorf(apperr.EMGVM, "Panic while executing operation %v", r)
		}
	}()

	if !entity.IsValidOperation(ref.Operation()) {
		return nil, apperr.Errorf(apperr.EFORBIDDEN, "invalid vm operation")
	}

	// Check if is enable CPU pool otherwise all operations are non blocking
	if vm.CPUsPoolService != nil {
		release, err := vm.CPUsPoolService.AcquireCore(ctx, ref)
		if err != nil {
			return nil, err
		}
		defer release()
	}

	// burn the max fuel consumed by the operation.
	if err := vm.FuelTank.Burn(vm.ctx, ref.MaxFuel()); err != nil {
		if err == service.ErrFuelTankNotEnough {
			vm.LogService.Info("Not enough fuel to execute operation, pause engine")
			vm.Pause()
		}
		return nil, err
	}

	startOpTime := time.Now()

	res, err = fn(ctx, ref)
	if err != nil {
		vm.LogService.Error(apperr.ErrorLog(err))
		return nil, err
	}

	if ref.WithRefuel() {

		// log the operation execution time.
		elapsed := time.Since(startOpTime)

		// calculate the fuel consumed effectively.
		effectiveFuelAmount := entity.FuelAmount(elapsed)

		// calculate the fuel saved.
		fuelRecovered := ref.MaxFuel() - effectiveFuelAmount

		// if fuel saved is greater than 0, refuel the tank.
		if fuelRecovered > 0 {
			if err := vm.FuelTank.Refuel(vm.ctx, fuelRecovered); err != nil {
				return nil, err
			}
		}
	}

	return res, nil
}

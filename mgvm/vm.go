package mgvm

import (
	"context"
	"sync"
	"time"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.VmService = (*MusicGangVM)(nil)
var _ service.VmCallableService = (*MusicGangVM)(nil)

// vmFunc is a generic function callback executed by the vm.
type vmFunc func(ctx context.Context, ref service.VmCallable) (interface{}, error)

// MusicGangVM is a virtual machine for the Mg language(nodeJS for now).
type MusicGangVM struct {
	ctx    context.Context
	cancel context.CancelFunc

	*sync.Cond

	LogService    service.LogService
	EngineService service.EngineService
	FuelTank      service.FuelTankService
	FuelStation   service.FuelStationService

	AuthManagmentService     service.AuthManagmentService
	ContractManagmentService service.ContractManagmentService
	UserManagmentService     service.UserManagmentService
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
	}
}

// Run starts the vm.
func (vm *MusicGangVM) Run() error {
	if err := vm.FuelStation.ResumeRefueling(vm.ctx); err != nil {
		return err
	}
	if err := vm.Resume(); err != nil {
		return err
	}
	go func() {
		errChan := make(chan error, 1)
		infoChan := make(chan string, 1)

		go vm.meter(infoChan, errChan)

		for {
			select {
			case <-vm.ctx.Done():
				return
			case err := <-errChan:
				vm.LogService.ReportError(vm.ctx, err)
			case info := <-infoChan:
				vm.LogService.ReportInfo(vm.ctx, info)
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
	if err := vm.EngineService.Stop(); err != nil {
		return err
	}
	if err := vm.FuelStation.StopRefueling(vm.ctx); err != nil {
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
func (vm *MusicGangVM) State() entity.State {
	return vm.EngineService.State()
}

// Stats returns the stats of fuel tank usage.
func (vm *MusicGangVM) Stats(ctx context.Context) (*entity.FuelStat, error) {
	return vm.FuelTank.Stats(ctx)
}

// Stop stops the engine.
// Delegates to the engine service.
func (vm *MusicGangVM) Stop() error {
	return vm.EngineService.Stop()
}

// makeOperations executes the given operations.
func (vm *MusicGangVM) makeOperations(ctx context.Context, ref service.VmCallable, fn vmFunc) (res interface{}, err error) {
	select {
	case <-ctx.Done():
		return nil, apperr.Errorf(apperr.EMGVM, "Timeout while executing contract")
	default:
		func() {
			vm.L.Lock()
			for !vm.IsRunning() {
				vm.LogService.ReportInfo(vm.ctx, "Wait for engine to resume")
				vm.Wait()
			}
			vm.L.Unlock()
		}()
	}

	defer func() {
		// handle engine timeout or panic
		if r := recover(); r != nil {
			if r == service.EngineExecutionTimeoutPanic {
				err = apperr.Errorf(apperr.EMGVM, "Timeout while executing contract")
				return
			}
			err = apperr.Errorf(apperr.EMGVM, "Panic while executing contract")
		}
	}()

	if !entity.IsValidOperation(ref.Operation()) {
		return nil, apperr.Errorf(apperr.EFORBIDDEN, "invalid vm operation")
	}

	// burn the max fuel consumed by the contract.
	if err := vm.FuelTank.Burn(vm.ctx, ref.MaxFuel()); err != nil {
		if err == service.ErrFuelTankNotEnough {
			vm.LogService.ReportInfo(vm.ctx, "Not enough fuel to execute contract, pause engine")
			vm.Pause()
		}
		return nil, err
	}

	startOpTime := time.Now()

	res, err = fn(ctx, ref)
	if err != nil {
		vm.LogService.ReportError(vm.ctx, err)
		return nil, err
	}

	if ref.WithRefuel() {

		// log the contract execution time.
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

// meter measures the fuel consumption of the engine.
func (vm *MusicGangVM) meter(infoChan chan<- string, errChan chan<- error) {

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	loop := true

	for loop {

		func() {

			defer func() {
				if r := recover(); r != nil {
					vm.LogService.ReportPanic(vm.ctx, r)
				}
			}()

			select {
			case <-vm.ctx.Done():
				loop = false
				return
			case <-ticker.C:
			}

			if vm.State() == entity.StatePaused {
				if fuel, err := vm.FuelTank.Fuel(vm.ctx); err != nil {
					errChan <- err
				} else if float64(fuel) <= float64(entity.FuelTankCapacity)*0.65 {
					vm.Resume()
					infoChan <- "Resume engine due to reaching safe fuel level"
				}
			} else if vm.State() == entity.StateRunning {
				if fuel, err := vm.FuelTank.Fuel(vm.ctx); err != nil {
					errChan <- err
				} else if float64(fuel) >= float64(entity.FuelTankCapacity)*0.95 {
					vm.Pause()
					infoChan <- "Pause engine due to excessive fuel consumption"
				}
			}
		}()
	}
}

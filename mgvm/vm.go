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
var _ service.FuelMeterService = (*MusicGangVM)(nil)

// MusicGangVM is a virtual machine for the Mg language(nodeJS for now).
type MusicGangVM struct {
	ctx    context.Context
	cancel context.CancelFunc

	*sync.Cond

	LogService    service.LogService
	EngineService service.EngineService
	FuelTank      service.FuelTankService
	FuelStation   service.FuelStationService
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
	vm.FuelStation.ResumeRefueling(vm.ctx)
	vm.Resume()
	go func() {

		for {

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)

			vm.ExecContract(ctx, &service.ContractCall{
				Contract: &entity.Contract{
					MaxFuel: entity.FuelLongActionAmount,
					LastRevision: &entity.Revision{
						Code: `
							function sum(a, b) {
								return a+b;
							}
							var result = sum(1, 2);
						`,
					},
				},
			})

			cancel()

			time.Sleep(100 * time.Millisecond)
		}
	}()

	go vm.meter()

	return nil
}

// Close closes the vm.
func (mg *MusicGangVM) Close() error {
	mg.cancel()
	mg.EngineService.Stop()
	mg.FuelStation.StopRefueling(mg.ctx)
	return nil
}

// ExecContract executes the contract.
// This func is a wrapper for the Engine.ExecContract.
func (vm *MusicGangVM) ExecContract(ctx context.Context, contractRef *service.ContractCall) (res interface{}, err error) {

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
			if r == EngineExecutionTimeoutPanic {
				err = apperr.Errorf(apperr.EMGVM, "Timeout while executing contract")
				return
			}
			err = apperr.Errorf(apperr.EMGVM, "Panic while executing contract")
		}
	}()

	// burn the max fuel consumed by the contract.
	if err := vm.FuelTank.Burn(vm.ctx, contractRef.Contract.MaxFuel); err != nil {
		if err == service.ErrFuelTankNotEnough {
			vm.LogService.ReportInfo(vm.ctx, "Not enough fuel to execute contract, pause engine")
			vm.Pause()
		}
		return nil, err
	}

	startContractTime := time.Now()

	// pass the contract to the engine.
	res, err = vm.EngineService.ExecContract(ctx, contractRef)
	if err != nil {
		vm.LogService.ReportError(vm.ctx, err)
		return nil, err
	}

	// log the contract execution time.
	elapsed := time.Since(startContractTime)

	// calculate the fuel consumed effectively.
	effectiveFuelAmount := entity.FuelAmount(elapsed)

	// calculate the fuel saved.
	fuelRecovered := contractRef.Contract.MaxFuel - effectiveFuelAmount

	// if fuel saved is greater than 0, refuel the tank.
	if fuelRecovered > 0 {
		if err := vm.FuelTank.Refuel(vm.ctx, fuelRecovered); err != nil {
			return nil, err
		}
	}

	return res, nil
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
func (vm *MusicGangVM) State() service.State {
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

// meter measures the fuel consumption of the engine.
func (vm *MusicGangVM) meter() {

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {

		func() {

			defer func() {
				if r := recover(); r != nil && vm.ctx.Err() == nil {
					vm.LogService.ReportPanic(vm.ctx, r)
				}
			}()

			select {
			case <-vm.ctx.Done():
				return
			case <-ticker.C:
			}

			if vm.State() == service.StatePaused {
				if fuel, err := vm.FuelTank.Fuel(vm.ctx); err != nil {
					vm.LogService.ReportError(vm.ctx, err)
				} else if float64(fuel) <= float64(entity.FuelTankCapacity)*0.65 {
					vm.Resume()
					vm.LogService.ReportInfo(vm.ctx, "Resume engine due to reaching safe fuel level")
				}
			} else if vm.State() == service.StateRunning {
				if fuel, err := vm.FuelTank.Fuel(vm.ctx); err != nil {
					vm.LogService.ReportError(vm.ctx, err)
				} else if float64(fuel) >= float64(entity.FuelTankCapacity)*0.95 {
					vm.Pause()
					vm.LogService.ReportInfo(vm.ctx, "Pause engine due to excessive fuel consumption")
				}
			}
		}()
	}
}

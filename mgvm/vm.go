package mgvm

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/robertkrimen/otto"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/app/util"
)

// MusicGangVM is a virtual machine for the Mg language(nodeJS for now).
type MusicGangVM struct {
	ctx    context.Context
	cancel context.CancelFunc

	state State

	LogService  service.LogService
	FuelTank    service.FuelTankService
	FuelStation service.FuelStationService
}

// MusicGangVM creates a new MusicGangVM.
// It should be called only once.
func NewMusicGangVM() *MusicGangVM {
	ctx := app.NewContextWithTags(context.Background(), []string{app.ContextTagMGVM})
	ctx, cancel := context.WithCancel(ctx)
	return &MusicGangVM{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (vm *MusicGangVM) Run() error {
	vm.Resume()
	vm.FuelStation.ResumeRefueling(vm.ctx)

	go func() {

		for {

			if err := vm.ExecAction(NewAction(vm.ctx, &entity.Contract{
				MaxFuel: entity.Fuel(3000),
			})); err != nil {
				vm.LogService.ReportError(vm.ctx, err)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	return nil
}

func (mg *MusicGangVM) Close() error {
	mg.FuelStation.StopRefueling(mg.ctx)
	mg.cancel()
	atomic.StoreInt32((*int32)(&mg.state), int32(StateClosed))
	return nil
}

func (mg *MusicGangVM) Pause() {
	atomic.StoreInt32((*int32)(&mg.state), int32(StatePaused))
}

func (mg *MusicGangVM) Resume() {
	atomic.StoreInt32((*int32)(&mg.state), int32(StateRunning))
}

func (mg *MusicGangVM) IsRunning() bool {
	return atomic.LoadInt32((*int32)(&mg.state)) == int32(StateRunning)
}

func (vm *MusicGangVM) ExecAction(action *Action) (err error) {

	defer func() {
		if r := recover(); r != nil {
			if r == "timeout" {
				err = apperr.Errorf(apperr.EMGVM, "Timeout while executing action")
				return
			}

			err = apperr.Errorf(apperr.EMGVM, "Panic while executing action")
		}
	}()

	if err := vm.FuelTank.Burn(vm.ctx, action.ContractMG.MaxFuel); err != nil {
		return err
	}

	timeoutTicker := time.NewTicker(entity.MaxExecutionTime)
	defer timeoutTicker.Stop()

	ottoVm := otto.New()
	ottoVm.Interrupt = make(chan func(), 1)

	go func() {
		<-timeoutTicker.C
		ottoVm.Interrupt <- func() {
			panic("timeout")
		}
	}()

	startActionTime := time.Now()

	script, err := ottoVm.Compile("", `
		function sum(a, b) {
			return a+b;
		}
		var result = sum(1, 2);
	`)
	if err != nil {
		action.res <- util.Err(apperr.Errorf(apperr.EINTERNAL, "Error compiling script: %s", err))
		return err
	}

	_, err = ottoVm.Run(script)
	close(ottoVm.Interrupt)

	if err != nil {
		action.res <- util.Err(apperr.Errorf(apperr.EINTERNAL, "Error while executing action: %s", err.Error()))
		return err
	}

	value, err := ottoVm.Get("result")
	if err != nil {
		action.res <- util.Err(apperr.Errorf(apperr.EINTERNAL, "Error while retrieving result: %s", err.Error()))
		return err
	}

	str, err := value.ToString()
	if err != nil {
		action.res <- util.Err(apperr.Errorf(apperr.EINTERNAL, "Error while parsing action result: %s", err.Error()))
		return err
	}

	elapsed := time.Since(startActionTime)

	effectiveFuelAmount := entity.FuelAmount(elapsed)

	fuelRecovered := action.ContractMG.MaxFuel - effectiveFuelAmount

	if fuelRecovered > 0 {
		if err := vm.FuelTank.Refuel(vm.ctx, fuelRecovered); err != nil {
			return err
		}
	}

	action.res <- util.Ok(str)

	return nil
}

type Action struct {
	ctx        context.Context
	res        chan util.Result
	ContractMG *entity.Contract
}

func NewAction(ctx context.Context, contract *entity.Contract) *Action {
	return &Action{
		ctx:        ctx,
		res:        make(chan util.Result, 1),
		ContractMG: contract,
	}
}

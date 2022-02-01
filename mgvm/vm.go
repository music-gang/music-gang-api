package mgvm

import (
	"context"
	"fmt"
	"sync"
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

	running     int32
	actionsChan chan *Action
	streamCtrl  chan bool

	LogService  service.LogService
	FuelTank    service.FuelTankService
	FuelStation service.FuelStationService
	Scheduler   *Scheduler
}

// MusicGangVM creates a new MusicGangVM.
// It should be called only once.
func NewMusicGangVM() *MusicGangVM {
	ctx := app.NewContextWithTags(context.Background(), []string{app.ContextTagMGVM})
	ctx, cancel := context.WithCancel(ctx)
	return &MusicGangVM{
		ctx:         ctx,
		cancel:      cancel,
		actionsChan: make(chan *Action, 10),
		streamCtrl:  make(chan bool, 1),
	}
}

func (vm *MusicGangVM) Run() error {
	vm.Resume()
	go vm.Scheduler.StreamActions(vm.ctx, vm.actionsChan, vm.streamCtrl)
	vm.FuelStation.ResumeRefueling(vm.ctx)
	go vm.ReadActions()
	go vm.FuelChecker()

	go func() {

		for {

			vm.ExecAction(NewAction(vm.ctx, nil))

			time.Sleep(100 * time.Millisecond)
		}
	}()

	return nil
}

func (mg *MusicGangVM) Close() error {
	mg.FuelStation.StopRefueling(mg.ctx)
	mg.cancel()
	return nil
}

func (mg *MusicGangVM) Pause() {
	atomic.StoreInt32(&mg.running, 0)
	// pause stream to actionsChan
	mg.streamCtrl <- false
}

func (mg *MusicGangVM) Resume() {
	atomic.StoreInt32(&mg.running, 1)
	// open stream to actionsChan
	mg.streamCtrl <- true
}

func (mg *MusicGangVM) IsRunning() bool {
	return atomic.LoadInt32(&mg.running) == 1
}

func (vm *MusicGangVM) FuelChecker() {

	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {

		func() {

			defer func() {
				if r := recover(); r != nil {
					vm.LogService.ReportPanic(vm.ctx, r)
				}
			}()

			if currentFuel, err := vm.FuelTank.Fuel(vm.ctx); err != nil {
				vm.LogService.ReportError(vm.ctx, err)
			} else if float64(currentFuel)/float64(entity.FuelTankCapacity) > 0.95 {
				vm.LogService.ReportWarning(vm.ctx, "Fuel tank is above 95%, stopping the vm")
				vm.Pause()
			}
		}()
	}
}

func (vm *MusicGangVM) ReadActions() {

	for {

		func() {

			defer func() {
				if r := recover(); r != nil {
					vm.LogService.ReportPanic(vm.ctx, r)
				}
			}()

			if vm.IsRunning() {

				action := <-vm.actionsChan

				go func(a *Action) {

					defer func() {
						if r := recover(); r != nil {
							vm.LogService.ReportPanic(vm.ctx, r)
						}
					}()

					if err := vm.ExecAction(action); err != nil {
						vm.LogService.ReportFatal(vm.ctx, err)
					}
				}(action)
			}

			time.Sleep(10 * time.Millisecond)
		}()
	}
}

func (vm *MusicGangVM) ExecAction(action *Action) error {

	defer func() {
		if r := recover(); r != nil {
			if r == "timeout" {
				vm.LogService.ReportWarning(vm.ctx, "Timeout while executing action")
				return
			}
			if r == "fuel-out" {
				vm.LogService.ReportWarning(vm.ctx, "Fuel out while executing action")
				return
			}
			if r == "halt" {
				vm.LogService.ReportWarning(vm.ctx, "HALT!")
			}
			panic(r)
		}
	}()

	ticker := time.NewTicker(200 * time.Millisecond)
	timeoutTicker := time.NewTicker(10 * time.Second)

	defer ticker.Stop()
	defer timeoutTicker.Stop()

	ottoVm := otto.New()
	ottoVm.Interrupt = make(chan func(), 1)

	go func() {
		lastFuelCheck := time.Now()
		for {
			select {
			case <-ticker.C:
				fuelConsumed := entity.FuelAmount(time.Since(lastFuelCheck))
				lastFuelCheck = time.Now()
				if err := vm.FuelTank.Burn(vm.ctx, fuelConsumed); err != nil {
					ottoVm.Interrupt <- func() { panic("fuel-out") }
				}
			case <-timeoutTicker.C:
				ottoVm.Interrupt <- func() { panic("timeout") }
			}
		}

	}()

	script, err := ottoVm.Compile("", `
		function sum(a, b) {
			return a+b;
		}
	`)
	if err != nil {
		action.res <- util.Err(apperr.Errorf(apperr.EINTERNAL, "Error compiling script: %s", err))
		vm.LogService.ReportWarning(vm.ctx, fmt.Sprintf("Error compiling script: %s", err.Error()))
		return err
	}

	_, err = ottoVm.Run(script)

	if err != nil {
		action.res <- util.Err(apperr.Errorf(apperr.EINTERNAL, "Error while executing action: %s", err.Error()))
		vm.LogService.ReportWarning(vm.ctx, fmt.Sprintf("Error while executing action: %s", err.Error()))
		return err
	}

	value, err := ottoVm.Get("result")
	if err != nil {
		action.res <- util.Err(apperr.Errorf(apperr.EINTERNAL, "Error while retrieving result: %s", err.Error()))
		vm.LogService.ReportWarning(vm.ctx, fmt.Sprintf("Error while retrieving result: %s", err.Error()))
		return err
	}

	str, err := value.ToString()
	if err != nil {
		action.res <- util.Err(apperr.Errorf(apperr.EINTERNAL, "Error while parsing action result: %s", err.Error()))
		vm.LogService.ReportWarning(vm.ctx, fmt.Sprintf("Error while executing action: %s", err.Error()))
		return err
	}

	println(str)

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

type Scheduler struct {
	muxQ  sync.Mutex
	queue []*Action
}

func (s *Scheduler) Push(action *Action) <-chan util.Result {
	s.muxQ.Lock()
	defer s.muxQ.Unlock()
	s.queue = append(s.queue, action)
	return action.res
}

func (s *Scheduler) StreamActions(ctx context.Context, actionsChan chan<- *Action, streamControl <-chan bool) {

	streamOpen := true

	for {

		select {
		case <-ctx.Done():
			return
		case sc := <-streamControl:
			streamOpen = sc
		default:
		}

		if streamOpen {
			// remove first action from queue
			s.muxQ.Lock()
			if len(s.queue) > 0 {

				action := s.queue[0]
				s.queue = s.queue[1:]

				// send action to actionsChan
				actionsChan <- action
			}
			s.muxQ.Unlock()
		}

		time.Sleep(10 * time.Millisecond)
	}
}

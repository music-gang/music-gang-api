package mgvm

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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

	LogService service.LogService
	FuelTank   service.FuelTankService
	Scheduler  *Scheduler
}

// MusicGangVM creates a new MusicGangVM.
// It should be called only once.
func NewMusicGangVM() *MusicGangVM {
	return &MusicGangVM{
		ctx:         context.Background(),
		actionsChan: make(chan *Action, 10),
		streamCtrl:  make(chan bool, 1),
	}
}

func (mg *MusicGangVM) Run(ctx context.Context) error {
	mg.ctx = ctx

	mg.Resume()

	go mg.Scheduler.StreamActions(mg.ctx, mg.actionsChan, mg.streamCtrl)
	go mg.AutoRefuel()
	go mg.ReadActions()
	go mg.FuelChecker()

	go func() {
		for {

			mg.Scheduler.Push(NewAction(mg.ctx, nil))

			time.Sleep(900 * time.Millisecond)
		}

	}()

	return nil
}

func (mg *MusicGangVM) Close() error {
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

func (vm *MusicGangVM) AutoRefuel() {

	tickerToRefuel := time.NewTicker(1 * time.Second)
	for range tickerToRefuel.C {

		func() {

			defer func() {
				if r := recover(); r != nil {
					vm.LogService.ReportPanic(vm.ctx, r)
				}
			}()

			if err := vm.FuelTank.Refuel(vm.ctx, entity.FuelRefillAmount); err != nil {
				vm.LogService.ReportFatal(vm.ctx, err)
			}

			fuel, err := vm.FuelTank.Fuel(vm.ctx)
			if err != nil {
				vm.LogService.ReportError(vm.ctx, err)
			}
			fmt.Printf("Cap: %d\n", entity.FuelTankCapacity)
			fmt.Printf("Fuel: %d\n", fuel)
			if float64(fuel)/float64(entity.FuelTankCapacity) < 0.65 {
				vm.Resume()
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

	done := make(chan bool, 1)
	ticker := time.NewTicker(200 * time.Millisecond)
	timeoutTicker := time.NewTicker(10 * time.Second)

	defer ticker.Stop()
	defer timeoutTicker.Stop()
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
			vm.LogService.ReportPanic(vm.ctx, r)
		}
	}()

	lastFuelCheck := time.Now()

	go func(a *Action) {

		time.Sleep(600 * time.Millisecond)

		// run contract

		println("WORKING")

		a.res <- util.Ok("ok")

		done <- true
		close(done)

	}(action)

loop:
	for {
		select {
		case <-ticker.C:
			fuelConsumed := entity.FuelAmount(time.Since(lastFuelCheck))
			lastFuelCheck = time.Now()
			if err := vm.FuelTank.Burn(vm.ctx, fuelConsumed); err != nil {
				panic("fuel-out")
			}
		case <-timeoutTicker.C:
			panic("timeout")
		case <-done:
			break loop
		}
	}

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

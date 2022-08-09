package mgvm_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/mgvm"
	"github.com/music-gang/music-gang-api/mock"
)

func TestVm_Run(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.FuelStation = &mock.FuelStationService{
			ResumeRefuelingFn: func(ctx context.Context) error {
				return nil
			},
			IsRunningFn: func() bool {
				return true
			},
			StopRefuelingFn: func(ctx context.Context) error {
				return nil
			},
		}

		isRunning := false

		vm.FuelTank = &mock.FuelTankServiceNoOp{}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return isRunning
			},
			PauseFn: func() error {
				isRunning = false
				return nil
			},
			ResumeFn: func() error {
				isRunning = true
				return nil
			},
			StateFn: func() entity.VmState {
				return entity.StateRunning
			},
			StopFn: func() error {
				isRunning = false
				return nil
			},
			ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
				return nil, nil
			},
		}
		vm.LogService = &mock.LoggerNoOp{}

		if err := vm.Run(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if !vm.IsRunning() || vm.State() != entity.StateRunning {
			t.Errorf("VM is not running")
		}

		if err := vm.Pause(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if vm.IsRunning() {
			t.Errorf("VM is running while paused")
		}

		if err := vm.Stop(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})

	t.Run("ResumeRefuelingErr", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.FuelStation = &mock.FuelStationService{
			ResumeRefuelingFn: func(ctx context.Context) error {
				return apperr.Errorf(apperr.EMGVM, "test")
			},
		}

		if err := vm.Run(); err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("EngineResumeErr", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.FuelStation = &mock.FuelStationService{
			ResumeRefuelingFn: func(ctx context.Context) error {
				return nil
			},
		}

		vm.EngineService = &mock.EngineService{
			ResumeFn: func() error {
				return apperr.Errorf(apperr.EMGVM, "test")
			},
		}

		if err := vm.Run(); err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("ListenOnInfoChan", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.FuelStation = &mock.FuelStationService{
			ResumeRefuelingFn: func(ctx context.Context) error {
				return nil
			},
			IsRunningFn: func() bool {
				return true
			},
			StopRefuelingFn: func(ctx context.Context) error {
				return nil
			},
		}

		isRunning := false

		currentFuel := entity.Fuel(0)

		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				atomic.AddUint64((*uint64)(&currentFuel), uint64(fuel))
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return entity.Fuel(atomic.LoadUint64((*uint64)(&currentFuel))), nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				atomic.AddUint64((*uint64)(&currentFuel), -uint64(fuelToRefill))
				return nil
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return isRunning
			},
			PauseFn: func() error {
				isRunning = false
				return nil
			},
			ResumeFn: func() error {
				isRunning = true
				return nil
			},
			StateFn: func() entity.VmState {
				return entity.StateRunning
			},
			StopFn: func() error {
				isRunning = false
				return nil
			},
			ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
				return nil, nil
			},
		}

		infoChan := make(chan string, 1)
		testIsFail := make(chan string, 1)

		vm.LogService = &mock.LoggerNoOp{
			InfoFn: func(msg string, ctx ...interface{}) {
				infoChan <- msg
			},
		}

		if err := vm.Run(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := vm.FuelTank.Burn(context.Background(), entity.Fuel(float64(entity.FuelTankCapacity)*0.96)); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		go func(tb testing.TB) {
			tb.Helper()
			time.Sleep(5 * time.Second)
			testIsFail <- "Expected info message"
		}(t)

		select {
		case <-infoChan:
		case err := <-testIsFail:
			t.Errorf("Unexpected error: %s", err)
		}
	})

	t.Run("ListenOnErrChan", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.FuelStation = &mock.FuelStationService{
			ResumeRefuelingFn: func(ctx context.Context) error {
				return nil
			},
			IsRunningFn: func() bool {
				return true
			},
			StopRefuelingFn: func(ctx context.Context) error {
				return nil
			},
		}

		isRunning := false

		currentFuel := entity.Fuel(0)

		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				atomic.AddUint64((*uint64)(&currentFuel), uint64(fuel))
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return entity.Fuel(0), apperr.Errorf(apperr.EMGVM, "test")
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				atomic.AddUint64((*uint64)(&currentFuel), -uint64(fuelToRefill))
				return nil
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return isRunning
			},
			PauseFn: func() error {
				isRunning = false
				return nil
			},
			ResumeFn: func() error {
				isRunning = true
				return nil
			},
			StateFn: func() entity.VmState {
				return entity.StateRunning
			},
			StopFn: func() error {
				isRunning = false
				return nil
			},
			ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
				return nil, nil
			},
		}

		ErrChan := make(chan error, 1)
		testIsFail := make(chan string, 1)

		vm.LogService = &mock.LoggerNoOp{
			ErrorFn: func(msg string, ctx ...interface{}) {
				ErrChan <- fmt.Errorf(msg, ctx...)
			},
		}

		if err := vm.Run(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := vm.FuelTank.Burn(context.Background(), entity.Fuel(float64(entity.FuelTankCapacity)*0.96)); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		go func(tb testing.TB) {
			tb.Helper()
			time.Sleep(5 * time.Second)
			testIsFail <- "Expected info message"
		}(t)

		select {
		case <-ErrChan:
		case err := <-testIsFail:
			t.Errorf("Unexpected error: %s", err)
		}
	})
}

func TestVm_Close(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.FuelStation = &mock.FuelStationService{
			ResumeRefuelingFn: func(ctx context.Context) error {
				return nil
			},
			IsRunningFn: func() bool {
				return true
			},
			StopRefuelingFn: func(ctx context.Context) error {
				return nil
			},
		}
		vm.FuelTank = &mock.FuelTankServiceNoOp{}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return true
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				return nil
			},
			StateFn: func() entity.VmState {
				return entity.StateRunning
			},
			StopFn: func() error {
				return nil
			},
			ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
				return nil, nil
			},
		}
		vm.LogService = &mock.LoggerNoOp{}

		if err := vm.Run(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := vm.Close(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})

	t.Run("StopRefuelingErr", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.FuelStation = &mock.FuelStationService{
			ResumeRefuelingFn: func(ctx context.Context) error {
				return nil
			},
			IsRunningFn: func() bool {
				return true
			},
			StopRefuelingFn: func(ctx context.Context) error {
				return apperr.Errorf(apperr.EMGVM, "test")
			},
		}
		vm.FuelTank = &mock.FuelTankServiceNoOp{}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return true
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				return nil
			},
			StateFn: func() entity.VmState {
				return entity.StateRunning
			},
			StopFn: func() error {
				return nil
			},
			ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
				return nil, nil
			},
		}
		vm.LogService = &mock.LoggerNoOp{}

		if err := vm.Run(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := vm.Close(); err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("EngineStopErr", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.FuelStation = &mock.FuelStationService{
			ResumeRefuelingFn: func(ctx context.Context) error {
				return nil
			},
			IsRunningFn: func() bool {
				return true
			},
			StopRefuelingFn: func(ctx context.Context) error {
				return nil
			},
		}
		vm.FuelTank = &mock.FuelTankServiceNoOp{}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return true
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				return nil
			},
			StateFn: func() entity.VmState {
				return entity.StateRunning
			},
			StopFn: func() error {
				return apperr.Errorf(apperr.EMGVM, "test")
			},
			ExecContractFn: func(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {
				return nil, nil
			},
		}
		vm.LogService = &mock.LoggerNoOp{}

		if err := vm.Run(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := vm.Close(); err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("EngineStateErr", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.EngineService = &mock.EngineService{
			StateFn: func() entity.VmState {
				return entity.StateInitializing
			},
		}
		if err := vm.Close(); err == nil {
			t.Errorf("Expected error")
		}

		vm.EngineService = &mock.EngineService{
			StateFn: func() entity.VmState {
				return entity.StateStopped
			},
		}
		if err := vm.Close(); err == nil {
			t.Errorf("Expected error")
		}
	})
}

func TestVm_Meter(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		ctx := context.Background()
		currentFuel := entity.Fuel(0)
		state := entity.StateInitializing
		muxState := &sync.Mutex{}

		vm.LogService = &mock.LoggerNoOp{}

		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				atomic.AddUint64((*uint64)(&currentFuel), uint64(fuel))
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return entity.Fuel(atomic.LoadUint64((*uint64)(&currentFuel))), nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				atomic.AddUint64((*uint64)(&currentFuel), -uint64(fuelToRefill))
				return nil
			},
		}

		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				muxState.Lock()
				defer muxState.Unlock()
				return state == entity.StateRunning
			},
			PauseFn: func() error {
				muxState.Lock()
				defer muxState.Unlock()
				state = entity.StatePaused
				return nil
			},
			ResumeFn: func() error {
				muxState.Lock()
				defer muxState.Unlock()
				state = entity.StateRunning
				return nil
			},
			StateFn: func() entity.VmState {
				muxState.Lock()
				defer muxState.Unlock()
				return state
			},
			StopFn: func() error {
				muxState.Lock()
				defer muxState.Unlock()
				state = entity.StateStopped
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		errChan := make(chan error, 1)
		infoChan := make(chan string, 1)

		go vm.Meter(infoChan, errChan)

		if err := vm.FuelTank.Burn(ctx, entity.Fuel(float64(entity.FuelTankCapacity)*0.96)); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		time.Sleep(time.Second)

		if vm.State() != entity.StatePaused {
			t.Errorf("Unexpected state after high vm usage, got: %s, want: %s", vm.State(), entity.StatePaused)
		}

		if err := vm.FuelTank.Refuel(ctx, entity.Fuel(float64(entity.FuelTankCapacity)*0.96)); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		time.Sleep(time.Second)

		if vm.State() != entity.StateRunning {
			t.Errorf("Unexpected state after low vm usage, got: %s, want: %s", vm.State(), entity.StateRunning)
		}
	})

	t.Run("ContextCanel", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.Cancel()

		errChan := make(chan error, 1)
		infoChan := make(chan string, 1)

		go vm.Meter(infoChan, errChan)

		time.Sleep(time.Second)
	})

	t.Run("FuelTankErrOnPausedState", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.LogService = &mock.LoggerNoOp{}
		vm.EngineService = &mock.EngineService{
			StateFn: func() entity.VmState {
				return entity.StatePaused
			},
		}
		vm.FuelTank = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return 0, apperr.Errorf(apperr.EMGVM, "test")
			},
		}

		errChan := make(chan error, 1)
		infoChan := make(chan string, 1)

		go vm.Meter(infoChan, errChan)

		time.Sleep(time.Second)

		shouldFail := true
		shouldFailLock := &sync.Mutex{}

		go func(tb testing.TB) {
			time.Sleep(5 * time.Second)
			shouldFailLock.Lock()
			defer shouldFailLock.Unlock()
			if shouldFail {
				tb.Fatal("Fuel tank error not received")
			}
		}(t)

		err := <-errChan
		if err == nil {
			t.Fatal("Fuel tank error not received")
		} else if code := apperr.ErrorCode(err); code != apperr.EMGVM {
			t.Errorf("Unexpected error code, got: %s, want: %s", code, apperr.EMGVM)
		}

		shouldFailLock.Lock()
		shouldFail = false
		shouldFailLock.Unlock()
	})

	t.Run("FuelTankErrOnRunningState", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.LogService = &mock.LoggerNoOp{}
		vm.EngineService = &mock.EngineService{
			StateFn: func() entity.VmState {
				return entity.StateRunning
			},
		}
		vm.FuelTank = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return 0, apperr.Errorf(apperr.EMGVM, "test")
			},
		}

		errChan := make(chan error, 1)
		infoChan := make(chan string, 1)

		go vm.Meter(infoChan, errChan)

		time.Sleep(time.Second)

		shouldFail := true
		shouldFailLock := &sync.Mutex{}

		go func(tb testing.TB) {
			time.Sleep(5 * time.Second)
			shouldFailLock.Lock()
			defer shouldFailLock.Unlock()
			if shouldFail {
				tb.Fatal("Fuel tank error not received")
			}
		}(t)

		err := <-errChan
		if err == nil {
			t.Fatal("Fuel tank error not received")
		} else if code := apperr.ErrorCode(err); code != apperr.EMGVM {
			t.Errorf("Unexpected error code, got: %s, want: %s", code, apperr.EMGVM)
		}

		shouldFailLock.Lock()
		shouldFail = false
		shouldFailLock.Unlock()
	})

	t.Run("Panic", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		vm.LogService = &mock.LoggerNoOp{}
		vm.EngineService = &mock.EngineService{
			StateFn: func() entity.VmState {
				return entity.StateRunning
			},
		}
		vm.FuelTank = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				panic("test")
			},
		}

		errChan := make(chan error, 1)
		infoChan := make(chan string, 1)

		go vm.Meter(infoChan, errChan)

		time.Sleep(time.Second)
	})
}

func TestVm_MakeOperation(t *testing.T) {

	t.Run("CPUsPool", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		releaseCoreCalled := false

		vm.CPUsPoolService = &mock.CPUsPoolService{
			AcquireCoreFn: func(ctx context.Context, call service.VmCallable) (release func(), err error) {
				return func() {
					releaseCoreCalled = true
				}, nil
			},
		}
		vm.LogService = &mock.LoggerNoOp{}
		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				atomic.StoreUint64((*uint64)(&currentFuel), uint64(fuel))
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return entity.Fuel(atomic.LoadUint64((*uint64)(&currentFuel))), nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				return nil
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.VmState(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.VmState {
				return entity.VmState(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Fatal(err)
		}

		var invalidOp entity.VmOperation = entity.VmOperationExecuteContract

		refCall := service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation: invalidOp,
		})

		if res, err := vm.MakeOperation(context.Background(), refCall, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
			return "ok", nil
		}); err != nil {
			t.Fatal(err)
		} else if res != "ok" {
			t.Errorf("Unexpected result, got: %s, want: %s", res, "ok")
		}

		if !releaseCoreCalled {
			t.Error("Release core not called")
		}
	})

	t.Run("CPUsPoolErr", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.CPUsPoolService = &mock.CPUsPoolService{
			AcquireCoreFn: func(ctx context.Context, call service.VmCallable) (release func(), err error) {
				return nil, apperr.Errorf(apperr.EMGVM, "internal error")
			},
		}
		vm.LogService = &mock.LoggerNoOp{}
		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				atomic.StoreUint64((*uint64)(&currentFuel), uint64(fuel))
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return entity.Fuel(atomic.LoadUint64((*uint64)(&currentFuel))), nil
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.VmState(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.VmState {
				return entity.VmState(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Fatal(err)
		}

		var invalidOp entity.VmOperation = entity.VmOperationExecuteContract

		refCall := service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation: invalidOp,
		})

		if _, err := vm.MakeOperation(context.Background(), refCall, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
			return "ok", nil
		}); err == nil {
			t.Error("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Unexpected error code, got: %s, want: %s", errCode, apperr.EMGVM)
		}
	})

	t.Run("InvalidOperation", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LoggerNoOp{}
		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				atomic.StoreUint64((*uint64)(&currentFuel), uint64(fuel))
				return nil
			},
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return entity.Fuel(atomic.LoadUint64((*uint64)(&currentFuel))), nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				panic("should not be called")
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.VmState(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.VmState {
				return entity.VmState(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Fatal(err)
		}

		var invalidOp entity.VmOperation = "invalid"

		refCall := service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation: invalidOp,
		})

		if _, err := vm.MakeOperation(context.Background(), refCall, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
			panic("should not be called")
		}); err == nil {
			t.Fatal("Expected error")
		} else if code := apperr.ErrorCode(err); code != apperr.EFORBIDDEN {
			t.Errorf("Unexpected error code, got: %s, want: %s", code, apperr.EFORBIDDEN)
		}
	})
}

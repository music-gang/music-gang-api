package mgvm_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/mgvm"
	"github.com/music-gang/music-gang-api/mock"
)

func TestVm_ExecContract(t *testing.T) {

	contract := &entity.Contract{
		MaxFuel: entity.FuelLongActionAmount,
		LastRevision: &entity.Revision{
			Code: `
					function sum(a, b) {
						return a+b;
					}
					var result = sum(1, 2);
				`,
		},
	}

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)
		refuelCalled := false

		vm.LogService = &mock.LogServiceNoOp{}
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
				refuelCalled = true
				return nil
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
			ExecContractFn: func(ctx context.Context, revision *entity.Revision) (res interface{}, err error) {
				return "contract executed", nil
			},
		}

		go func() {

			// simulate late start to mock the gorutine waiting for the engine to be running

			time.Sleep(time.Second)

			if err := vm.Resume(); err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
			}
		}()

		res, err := vm.ExecContract(context.Background(), contract.LastRevision)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if res != "contract executed" {
			t.Errorf("Unexpected result, got: %s, want: %s", res, "contract executed")
		}

		if !refuelCalled {
			t.Errorf("Refuel not called")
		}
	})

	t.Run("ContextCancelled", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		if _, err := vm.ExecContract(ctx, &entity.Revision{}); err == nil {
			t.Errorf("Expected error, got: %v", err)
		}
	})

	t.Run("WaitButContextCancelled", func(t *testing.T) {

	})

	t.Run("FuelTankErr", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing

		vm.LogService = &mock.LogServiceNoOp{}
		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return apperr.Errorf(apperr.EMGVM, "test")
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
			ExecContractFn: func(ctx context.Context, revision *entity.Revision) (res interface{}, err error) {
				return "contract executed", nil
			},
		}

		vm.EngineService.Resume()

		if _, err := vm.ExecContract(context.Background(), contract.LastRevision); err == nil {
			t.Errorf("Expected error, got: %v", err)
		}
	})

	t.Run("FuelTankNotEnough", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing

		vm.LogService = &mock.LogServiceNoOp{}
		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return service.ErrFuelTankNotEnough
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
			ExecContractFn: func(ctx context.Context, revision *entity.Revision) (res interface{}, err error) {
				return "contract executed", nil
			},
		}

		vm.EngineService.Resume()

		if _, err := vm.ExecContract(context.Background(), contract.LastRevision); err == nil {
			t.Errorf("Expected error, got: %v", err)
		} else if err != service.ErrFuelTankNotEnough {
			t.Errorf("Unexpected error, got: %v, want: %v", err, service.ErrFuelTankNotEnough)
		}
	})

	t.Run("RefuelErr", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing

		vm.LogService = &mock.LogServiceNoOp{}
		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return nil
			},
			RefuelFn: func(ctx context.Context, fuelToRefill entity.Fuel) error {
				return apperr.Errorf(apperr.EMGVM, "test")
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
			ExecContractFn: func(ctx context.Context, revision *entity.Revision) (res interface{}, err error) {
				return "contract executed", nil
			},
		}

		vm.EngineService.Resume()

		if _, err := vm.ExecContract(context.Background(), contract.LastRevision); err == nil {
			t.Errorf("Expected error, got: %v", err)
		} else if code := apperr.ErrorCode(err); code != apperr.EMGVM {
			t.Errorf("Unexpected error, got: %v, want: %v", code, apperr.EMGVM)
		}
	})

	t.Run("ExecContractErr", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing

		vm.LogService = &mock.LogServiceNoOp{}
		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return nil
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
			ExecContractFn: func(ctx context.Context, revision *entity.Revision) (res interface{}, err error) {
				return nil, apperr.Errorf(apperr.EMGVM, "test")
			},
		}

		vm.EngineService.Resume()

		if _, err := vm.ExecContract(context.Background(), contract.LastRevision); err == nil {
			t.Errorf("Expected error, got: %v", err)
		} else if code := apperr.ErrorCode(err); code != apperr.EMGVM {
			t.Errorf("Unexpected error, got: %v, want: %v", code, apperr.EMGVM)
		}
	})

	t.Run("ExecutionTimeout", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing

		vm.LogService = &mock.LogServiceNoOp{}
		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return nil
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
			ExecContractFn: func(ctx context.Context, revision *entity.Revision) (res interface{}, err error) {
				panic(service.EngineExecutionTimeoutPanic)
			},
		}

		vm.EngineService.Resume()

		if _, err := vm.ExecContract(context.Background(), contract.LastRevision); err == nil {
			t.Errorf("Expected error, got: %v", err)
		} else if code := apperr.ErrorCode(err); code != apperr.EMGVM {
			t.Errorf("Unexpected error, got: %v, want: %v", code, apperr.EMGVM)
		}
	})

	t.Run("GenericPanic", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing

		vm.LogService = &mock.LogServiceNoOp{}
		vm.FuelTank = &mock.FuelTankService{
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return nil
			},
		}
		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
			ExecContractFn: func(ctx context.Context, revision *entity.Revision) (res interface{}, err error) {
				panic("test")
			},
		}

		vm.EngineService.Resume()

		if _, err := vm.ExecContract(context.Background(), contract.LastRevision); err == nil {
			t.Errorf("Expected error, got: %v", err)
		} else if code := apperr.ErrorCode(err); code != apperr.EMGVM {
			t.Errorf("Unexpected error, got: %v, want: %v", code, apperr.EMGVM)
		}
	})
}

// All tests cases for the ExecContract method cover all possible scenarios inside makeOperations.
// So for other vm services I think it's not necessary repeat all tests cases for the ExecContract method.

func TestVm_CreateContract(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}
		vm.ContractManagmentService = &mock.ContractService{
			CreateContractFn: func(ctx context.Context, contract *entity.Contract) error {
				contract.ID = 1
				return nil
			},
		}

		go func() {

			// simulate late start to mock the gorutine waiting for the engine to be running

			time.Sleep(time.Second)

			if err := vm.Resume(); err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
			}
		}()

		contract := &entity.Contract{
			Name:    "test",
			MaxFuel: entity.FuelLongActionAmount,
		}

		err := vm.CreateContract(context.Background(), contract)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if contract.ID != 1 {
			t.Errorf("Unexpected contract ID: %d", contract.ID)
		}
	})
}

func TestVm_DeleteContract(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}
		vm.ContractManagmentService = &mock.ContractService{
			DeleteContractFn: func(ctx context.Context, id int64) error {
				return nil
			},
		}

		go func() {

			// simulate late start to mock the gorutine waiting for the engine to be running

			time.Sleep(time.Second)

			if err := vm.Resume(); err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
			}
		}()

		if err := vm.DeleteContract(context.Background(), 1); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})
}

func TestVm_MakeRevision(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}
		vm.ContractManagmentService = &mock.ContractService{
			MakeRevisionFn: func(ctx context.Context, revision *entity.Revision) error {
				revision.ID = 1
				return nil
			},
		}

		go func() {

			// simulate late start to mock the gorutine waiting for the engine to be running

			time.Sleep(time.Second)

			if err := vm.Resume(); err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
			}
		}()

		contract := &entity.Contract{
			MaxFuel: entity.FuelLongActionAmount,
			LastRevision: &entity.Revision{
				Code: `
					function sum(a, b) {
						return a+b;
					}
					var result = sum(1, 2);
				`,
			},
		}

		if err := vm.MakeRevision(context.Background(), contract.LastRevision); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if contract.LastRevision.ID != 1 {
			t.Errorf("Unexpected revision ID: %d", contract.LastRevision.ID)
		}
	})
}

func TestVm_Stats(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		now := time.Now()

		currentState := entity.StateInitializing

		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				panic("should not be called")
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		vm.FuelTank = &mock.FuelTankService{
			StatsFn: func(ctx context.Context) (*entity.FuelStat, error) {
				return &entity.FuelStat{
					FuelCapacity:    100,
					FuelUsed:        50,
					LastRefuelAmout: 5,
					LastRefuelAt:    now,
				}, nil
			},
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if stats, err := vm.Stats(context.Background()); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		} else if stats.FuelCapacity != 100 {
			t.Errorf("Unexpected stats, got: %d, want: %d", stats.FuelCapacity, 100)
		} else if stats.FuelUsed != 50 {
			t.Errorf("Unexpected stats, got: %d, want: %d", stats.FuelUsed, 50)
		} else if stats.LastRefuelAmout != 5 {
			t.Errorf("Unexpected stats, got: %d, want: %d", stats.LastRefuelAmout, 5)
		} else if stats.LastRefuelAt.Unix() != now.Unix() {
			t.Errorf("Unexpected stats, got: %s, want: %s", stats.LastRefuelAt, now)
		}
	})

	t.Run("ErrStats", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing

		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				panic("should not be called")
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		vm.FuelTank = &mock.FuelTankService{
			StatsFn: func(ctx context.Context) (*entity.FuelStat, error) {
				return nil, apperr.Errorf(apperr.EMGVM, "test error")
			},
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if _, err := vm.Stats(context.Background()); err == nil {
			t.Errorf("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Unexpected error got %s, want %s", errCode, apperr.EMGVM)
		}
	})

	t.Run("InvalidResult", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing

		vm.EngineService = &mock.EngineService{
			IsRunningFn: func() bool {
				panic("should not be called")
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		vm.FuelTank = &mock.FuelTankService{
			StatsFn: func(ctx context.Context) (*entity.FuelStat, error) {
				return nil, nil
			},
			BurnFn: func(ctx context.Context, fuel entity.Fuel) error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if _, err := vm.Stats(context.Background()); err == nil {
			t.Errorf("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Errorf("Unexpected error got %s, want %s", errCode, apperr.EINTERNAL)
		}
	})
}

func TestVm_UpdateContract(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		contract := &entity.Contract{
			ID:          1,
			Name:        "test",
			Description: "test",
			MaxFuel:     entity.FuelLongActionAmount,
			LastRevision: &entity.Revision{
				Code: `
					function sum(a, b) {
						return a+b;
					}
					var result = sum(1, 2);
				`,
			},
		}

		vm.ContractManagmentService = &mock.ContractService{
			UpdateContractFn: func(ctx context.Context, id int64, upd service.ContractUpdate) (*entity.Contract, error) {
				if upd.Name != nil {
					contract.Name = *upd.Name
				}
				if upd.Description != nil {
					contract.Description = *upd.Description
				}
				return contract, nil
			},
		}

		go func() {

			// simulate late start to mock the gorutine waiting for the engine to be running

			time.Sleep(time.Second)

			if err := vm.Resume(); err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
			}
		}()

		newContractName := "test-new"
		newContractDescription := "test-new"

		if _, err := vm.UpdateContract(context.Background(), contract.ID, service.ContractUpdate{
			Name:        &newContractName,
			Description: &newContractDescription,
		}); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})

	t.Run("ErrUpdate", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		vm.ContractManagmentService = &mock.ContractService{
			UpdateContractFn: func(ctx context.Context, id int64, upd service.ContractUpdate) (*entity.Contract, error) {
				return nil, apperr.Errorf(apperr.EINTERNAL, "test")
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		newContractName := "test-new"
		newContractDescription := "test-new"

		if _, err := vm.UpdateContract(context.Background(), 1, service.ContractUpdate{
			Name:        &newContractName,
			Description: &newContractDescription,
		}); err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("InvalidResult", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		vm.ContractManagmentService = &mock.ContractService{
			UpdateContractFn: func(ctx context.Context, id int64, upd service.ContractUpdate) (*entity.Contract, error) {
				return nil, nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		newContractName := "test-new"
		newContractDescription := "test-new"

		if _, err := vm.UpdateContract(context.Background(), 1, service.ContractUpdate{
			Name:        &newContractName,
			Description: &newContractDescription,
		}); err == nil {
			t.Errorf("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Errorf("Expected error code %s, got %s", apperr.EINTERNAL, errCode)
		}
	})
}

func TestVm_Auhenticate(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.AuthManagmentService = &mock.AuthService{
			AuthentcateFn: func(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {
				return &entity.Auth{
					ID:     1,
					UserID: 1,
					Source: entity.AuthSourceLocal,
					User: &entity.User{
						ID:   1,
						Name: "test",
					},
				}, nil
			},
		}

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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if auth, err := vm.Auhenticate(context.Background(), &entity.AuthUserOptions{}); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		} else if auth.ID != 1 {
			t.Errorf("Expected auth ID to be 1")
		} else if auth.UserID != 1 {
			t.Errorf("Expected user ID to be 1")
		} else if auth.Source != entity.AuthSourceLocal {
			t.Errorf("Expected auth source to be local")
		} else if auth.User.ID != auth.UserID {
			t.Errorf("Expected user ID to be equal to user ID")
		} else if auth.User.Name != "test" {
			t.Errorf("Expected user name to be test")
		}
	})

	t.Run("ErrAuthenticate", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.AuthManagmentService = &mock.AuthService{
			AuthentcateFn: func(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {
				return nil, apperr.Errorf(apperr.EMGVM, "test")
			},
		}

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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if _, err := vm.Auhenticate(context.Background(), &entity.AuthUserOptions{}); err == nil {
			t.Errorf("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Expected error code to be EMGVM")
		}
	})

	t.Run("InvalidResult", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.AuthManagmentService = &mock.AuthService{
			AuthentcateFn: func(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {
				return nil, nil
			},
		}

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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if auth, err := vm.Auhenticate(context.Background(), &entity.AuthUserOptions{}); err == nil {
			t.Errorf("Unexpected error: %s", err.Error())
		} else if auth != nil {
			t.Errorf("Expected auth to be nil")
		}
	})
}

func TestVm_CreateAuth(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.AuthManagmentService = &mock.AuthService{
			CreateAuthFn: func(ctx context.Context, auth *entity.Auth) error {
				auth.ID = 1
				return nil
			},
		}

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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		auth := &entity.Auth{
			UserID: 1,
			Source: entity.AuthSourceLocal,
		}

		if err := vm.CreateAuth(context.Background(), auth); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		} else if auth.ID != 1 {
			t.Errorf("Expected auth id to be 1")
		}
	})

	t.Run("ErrCreateAuth", func(t *testing.T) {

		t.Run("OK", func(t *testing.T) {

			vm := mgvm.NewMusicGangVM()

			currentState := entity.StateInitializing
			currentFuel := entity.Fuel(0)

			vm.LogService = &mock.LogServiceNoOp{}
			vm.AuthManagmentService = &mock.AuthService{
				CreateAuthFn: func(ctx context.Context, auth *entity.Auth) error {
					return apperr.Errorf(apperr.EMGVM, "test")
				},
			}

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
					return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
				},
				PauseFn: func() error {
					return nil
				},
				ResumeFn: func() error {
					atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
					return nil
				},
				StateFn: func() entity.State {
					return entity.State(atomic.LoadInt32((*int32)(&currentState)))
				},
				StopFn: func() error {
					return nil
				},
			}

			if err := vm.Resume(); err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
			}

			auth := &entity.Auth{
				UserID: 1,
				Source: entity.AuthSourceLocal,
			}

			if err := vm.CreateAuth(context.Background(), auth); err == nil {
				t.Errorf("Expected error")
			} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
				t.Errorf("Expected error code to be %s, got %s", apperr.EMGVM, errCode)
			}
		})
	})
}

func TestVm_CreateUser(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.UserManagmentService = &mock.UserService{
			CreateUserFn: func(ctx context.Context, user *entity.User) error {
				user.ID = 1
				return nil
			},
		}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		user := &entity.User{
			Name: "test",
		}

		if err := vm.CreateUser(context.Background(), user); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		} else if user.ID != 1 {
			t.Errorf("Expected user id to be 1")
		}
	})

	t.Run("ErrCreateUser", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.UserManagmentService = &mock.UserService{
			CreateUserFn: func(ctx context.Context, user *entity.User) error {
				return apperr.Errorf(apperr.EMGVM, "test")
			},
		}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		user := &entity.User{
			Name: "test",
		}

		if err := vm.CreateUser(context.Background(), user); err == nil {
			t.Errorf("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Expected error code to be %s, got %s", apperr.EMGVM, errCode)
		}
	})
}

func TestVm_DeleteAuth(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.AuthManagmentService = &mock.AuthService{
			DeleteAuthFn: func(ctx context.Context, id int64) error {
				return nil
			},
		}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := vm.DeleteAuth(context.Background(), 1); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})

	t.Run("ErrDeleteAuth", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.AuthManagmentService = &mock.AuthService{
			DeleteAuthFn: func(ctx context.Context, id int64) error {
				return apperr.Errorf(apperr.EMGVM, "test")
			},
		}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := vm.DeleteAuth(context.Background(), 1); err == nil {
			t.Errorf("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Expected error code to be %s, got %s", apperr.EMGVM, errCode)
		}
	})
}

func TestVm_DeleteUser(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.UserManagmentService = &mock.UserService{
			DeleteUserFn: func(ctx context.Context, id int64) error {
				return nil
			},
		}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := vm.DeleteUser(context.Background(), 1); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})

	t.Run("ErrDeleteAuth", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.LogService = &mock.LogServiceNoOp{}
		vm.UserManagmentService = &mock.UserService{
			DeleteUserFn: func(ctx context.Context, id int64) error {
				return apperr.Errorf(apperr.EMGVM, "test")
			},
		}
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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := vm.DeleteUser(context.Background(), 1); err == nil {
			t.Errorf("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Expected error code to be %s, got %s", apperr.EMGVM, errCode)
		}
	})
}

func TestVm_UpdateUser(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.UserManagmentService = &mock.UserService{
			UpdateUserFn: func(ctx context.Context, id int64, user service.UserUpdate) (*entity.User, error) {
				return &entity.User{
					ID:   1,
					Name: *user.Name,
				}, nil
			},
		}

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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		newName := "new name"

		upd := service.UserUpdate{
			Name: &newName,
		}

		if user, err := vm.UpdateUser(context.Background(), 1, upd); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		} else if user.Name != newName {
			t.Errorf("Expected name to be %s, got %s", newName, user.Name)
		}
	})

	t.Run("ErrUpdateUser", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.UserManagmentService = &mock.UserService{
			UpdateUserFn: func(ctx context.Context, id int64, user service.UserUpdate) (*entity.User, error) {
				return nil, apperr.Errorf(apperr.EMGVM, "test")
			},
		}

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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		newName := "new name"

		upd := service.UserUpdate{
			Name: &newName,
		}

		if _, err := vm.UpdateUser(context.Background(), 1, upd); err == nil {
			t.Errorf("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Expected error code to be %s, got %s", apperr.EMGVM, errCode)
		}
	})

	t.Run("InvalidResult", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		currentState := entity.StateInitializing
		currentFuel := entity.Fuel(0)

		vm.UserManagmentService = &mock.UserService{
			UpdateUserFn: func(ctx context.Context, id int64, user service.UserUpdate) (*entity.User, error) {
				return nil, nil
			},
		}

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
				return entity.State(atomic.LoadInt32((*int32)(&currentState))) == entity.StateRunning
			},
			PauseFn: func() error {
				return nil
			},
			ResumeFn: func() error {
				atomic.StoreInt32((*int32)(&currentState), int32(entity.StateRunning))
				return nil
			},
			StateFn: func() entity.State {
				return entity.State(atomic.LoadInt32((*int32)(&currentState)))
			},
			StopFn: func() error {
				return nil
			},
		}

		if err := vm.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		newName := "new name"

		upd := service.UserUpdate{
			Name: &newName,
		}

		if _, err := vm.UpdateUser(context.Background(), 1, upd); err == nil {
			t.Errorf("Expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EINTERNAL {
			t.Errorf("Expected error code to be %s, got %s", apperr.EINTERNAL, errCode)
		}
	})
}

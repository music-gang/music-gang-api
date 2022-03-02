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
		}
	})
}

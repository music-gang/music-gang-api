package mgvm_test

import (
	"context"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/mgvm"
)

func TestEngine_Resume(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})

	t.Run("AlreadyRunning", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := engine.Resume(); err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestEngine_Stop(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := engine.Stop(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})

	t.Run("AlreadyStopped", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := engine.Stop(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := engine.Stop(); err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestEngine_Pause(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := engine.Pause(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})

	t.Run("AlreadyPaused", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := engine.Pause(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := engine.Pause(); err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("AlreadyStopped", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Stop(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if err := engine.Pause(); err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestEngine_ExecContract(t *testing.T) {

	contractCall := &service.ContractCall{
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
	}

	t.Run("OK", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if res, err := engine.ExecContract(context.Background(), contractCall); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		} else if res == nil {
			t.Errorf("Expected response, got nil")
		} else if res.(string) != "3" {
			t.Errorf("Expected response, got %v", res)
		}
	})

	t.Run("EngineStopped", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Stop(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if _, err := engine.ExecContract(context.Background(), contractCall); err == nil {
			t.Errorf("Expected error, got nil")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Expected error code %s, got %s", apperr.EMGVM, errCode)
		}
	})

	t.Run("ContextCancelled", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		if _, err := engine.ExecContract(ctx, contractCall); err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("EngineExecutionTimeout", func(t *testing.T) {

		longContractCall := &service.ContractCall{
			Contract: &entity.Contract{
				MaxFuel: entity.FuelInstantActionAmount,
				LastRevision: &entity.Revision{
					Code: `
				var result = 0
				for (var i = 0; i < 1000000000; i++) {
					result = result + i
				}
				`,
				},
			},
		}

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic, got nil")
			}
		}()

		engine.ExecContract(context.Background(), longContractCall)
	})

	t.Run("ErrRuntime", func(t *testing.T) {

		failContractCall := &service.ContractCall{
			Contract: &entity.Contract{
				MaxFuel: entity.FuelInstantActionAmount,
				LastRevision: &entity.Revision{
					Code: `
					var result = 0
					for (var i = 0; i < 1000000000; i++) {
					`,
				},
			},
		}

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if _, err := engine.ExecContract(context.Background(), failContractCall); err == nil {
			t.Errorf("Expected error, got nil")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM {
			t.Errorf("Expected error code %s, got %s", apperr.EMGVM, errCode)
		}
	})
}

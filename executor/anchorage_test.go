package executor_test

import (
	"context"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/executor"
)

func TestAnchorageContractExecutor_ExecContract(t *testing.T) {

	code := `
		function sum(a, b) {
			return a+b;
		}
		var result = sum(1, 2);
	`

	contract := &entity.Contract{
		MaxFuel: entity.FuelLongActionAmount,
		LastRevision: &entity.Revision{
			Code:         code,
			CompiledCode: []byte(code),
		},
	}

	t.Run("OK", func(t *testing.T) {

		executor := executor.NewAnchorageContractExecutor()

		if res, err := executor.ExecContract(context.Background(), contract.LastRevision); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		} else if res == nil {
			t.Errorf("Expected response, got nil")
		} else if res.(string) != "3" {
			t.Errorf("Expected response, got %v", res)
		}
	})

	t.Run("ContextCancelled", func(t *testing.T) {

		executor := executor.NewAnchorageContractExecutor()

		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		if _, err := executor.ExecContract(ctx, contract.LastRevision); err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("EngineExecutionTimeout", func(t *testing.T) {

		code := `
			var result = 0
			for (var i = 0; i < 1000000000; i++) {
				result = result + i
			}
		`

		contract := &entity.Contract{
			MaxFuel: entity.FuelLongActionAmount,
			LastRevision: &entity.Revision{
				Code:         code,
				CompiledCode: []byte(code),
			},
		}

		executor := executor.NewAnchorageContractExecutor()

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic, got nil")
			}
		}()

		executor.ExecContract(context.Background(), contract.LastRevision)
	})

	t.Run("ErrRuntime", func(t *testing.T) {

		code := `
			var result = 0
			for (var i = 0; i < 1000000000; i++) {
		`

		contract := &entity.Contract{
			MaxFuel: entity.FuelLongActionAmount,
			LastRevision: &entity.Revision{
				Code:         code,
				CompiledCode: []byte(code),
			},
		}

		executor := executor.NewAnchorageContractExecutor()

		if _, err := executor.ExecContract(context.Background(), contract.LastRevision); err == nil {
			t.Errorf("Expected error, got nil")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EANCHORAGE {
			t.Errorf("Expected error code %s, got %s", apperr.EANCHORAGE, errCode)
		}
	})
}

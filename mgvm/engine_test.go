package mgvm_test

import (
	"context"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/mgvm"
	"github.com/music-gang/music-gang-api/mock"
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

	t.Run("OK", func(t *testing.T) {

		engine := mgvm.NewEngine()

		solution := "5"

		engine.Executor = &mock.ExecutorService{
			ExecContractFn: func(ctx context.Context, contractRef *service.ContractCall) (res interface{}, err error) {
				return solution, nil
			},
		}

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		res, err := engine.ExecContract(context.Background(), &service.ContractCall{})
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if res != solution {
			t.Errorf("Expected %s, got %s", solution, res)
		}
	})

	t.Run("EngineNotRunning", func(t *testing.T) {

		engine := mgvm.NewEngine()

		_, err := engine.ExecContract(context.Background(), &service.ContractCall{})
		if err == nil {
			t.Errorf("Expected error, got nil")
		} else if code := apperr.ErrorCode(err); code != apperr.EMGVM {
			t.Errorf("Expected error code %s, got %s", apperr.EMGVM, code)
		}
	})
}

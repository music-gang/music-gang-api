package mgvm_test

import (
	"context"
	"testing"

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

	t.Run("OK", func(t *testing.T) {

		engine := mgvm.NewEngine()

		if err := engine.Resume(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

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

		if res, err := engine.ExecContract(context.Background(), contractCall); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		} else if res == nil {
			t.Errorf("Expected response, got nil")
		} else if res.(string) != "3" {
			t.Errorf("Expected response, got %v", res)
		}
	})
}

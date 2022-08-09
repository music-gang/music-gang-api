package mgvm_test

import (
	"context"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/mgvm"
	"github.com/music-gang/music-gang-api/mock"
)

func TestFuelStation_ResumeRefueling(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel()

		fuelStation := mgvm.NewFuelStation()
		fuelStation.FuelRefillRate = time.Second
		fuelStation.FuelRefillAmount = entity.Fuel(0)

		fuelStation.LogService = &mock.LoggerNoOp{}
		fuelStation.FuelTankService = &mock.FuelTankServiceNoOp{}

		if err := fuelStation.ResumeRefueling(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		if !fuelStation.IsRunning() {
			t.Errorf("Expected running state to be true, got false")
		}

		if err := fuelStation.StopRefueling(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		if fuelStation.IsRunning() {
			t.Errorf("Expected running state to be false, got true")
		}
	})

	t.Run("ResumeRefuelErr", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel()

		fuelStation := mgvm.NewFuelStation()
		fuelStation.FuelRefillRate = 500 * time.Millisecond
		fuelStation.FuelRefillAmount = entity.Fuel(0)

		errChan := make(chan *apperr.Error, 1)
		timeout := time.NewTimer(2 * time.Second)

		fuelStation.LogService = &mock.LoggerNoOp{}
		fuelStation.FuelTankService = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return 0, nil
			},
			RefuelFn: func(ctx context.Context, amount entity.Fuel) error {
				errChan <- apperr.Errorf(apperr.EMGVM, "test error")
				return apperr.Errorf(apperr.EMGVM, "test error")
			},
		}

		if err := fuelStation.ResumeRefueling(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		select {
		case err := <-errChan:
			if err.Code != apperr.EMGVM {
				t.Errorf("Expected error code %v, got %v", apperr.EMGVM, err.Code)
			}
			timeout.Stop()
		case <-timeout.C:
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("ResumeRefuelPanic", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel()

		fuelStation := mgvm.NewFuelStation()
		fuelStation.FuelRefillRate = 500 * time.Millisecond
		fuelStation.FuelRefillAmount = entity.Fuel(0)

		fuelStation.LogService = &mock.LoggerNoOp{}
		fuelStation.FuelTankService = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return 0, nil
			},
			RefuelFn: func(ctx context.Context, amount entity.Fuel) error {
				panic("test panic")
			},
		}

		if err := fuelStation.ResumeRefueling(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(500 * time.Millisecond)
	})

	t.Run("ResumeRefuelAlreadyRunning", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel()

		fuelStation := mgvm.NewFuelStation()
		fuelStation.FuelRefillRate = 500 * time.Millisecond
		fuelStation.FuelRefillAmount = entity.Fuel(0)

		fuelStation.LogService = &mock.LoggerNoOp{}
		fuelStation.FuelTankService = &mock.FuelTankService{
			FuelFn: func(ctx context.Context) (entity.Fuel, error) {
				return 0, nil
			},
			RefuelFn: func(ctx context.Context, amount entity.Fuel) error {
				return nil
			},
		}

		if err := fuelStation.ResumeRefueling(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		if err := fuelStation.ResumeRefueling(ctx); err == nil {
			t.Errorf("Expected error, got none")
		}
	})
}

func TestFuelStation_StopRefueling(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel()

		fuelStation := mgvm.NewFuelStation()
		fuelStation.FuelRefillRate = 500 * time.Millisecond
		fuelStation.FuelRefillAmount = entity.Fuel(0)

		fuelStation.LogService = &mock.LoggerNoOp{}
		fuelStation.FuelTankService = &mock.FuelTankServiceNoOp{}

		if err := fuelStation.ResumeRefueling(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		if err := fuelStation.StopRefueling(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		if fuelStation.IsRunning() {
			t.Errorf("Expected running state to be false, got true")
		}
	})

	t.Run("AlreadyStopped", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel()

		fuelStation := mgvm.NewFuelStation()
		fuelStation.FuelRefillRate = 500 * time.Millisecond
		fuelStation.FuelRefillAmount = entity.Fuel(0)

		fuelStation.LogService = &mock.LoggerNoOp{}
		fuelStation.FuelTankService = &mock.FuelTankServiceNoOp{}

		if err := fuelStation.ResumeRefueling(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		if err := fuelStation.StopRefueling(ctx); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
		if err := fuelStation.StopRefueling(ctx); err == nil {
			t.Errorf("Expected error, got none")
		}
	})
}

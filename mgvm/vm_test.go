package mgvm_test

import (
	"context"
	"testing"

	"github.com/music-gang/music-gang-api/mgvm"
	"github.com/music-gang/music-gang-api/mock"
)

func TestVm_Run(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		vm := mgvm.NewMusicGangVM()

		fuelStation := &mock.FuelStationService{
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

		vm.FuelStation = fuelStation

		if err := vm.Run(); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	})
}

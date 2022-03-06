package mgvm

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

func SwitchToLocalFuel() {
	useRemoteFuel = false
}

func SwitchToRemoteFuel() {
	useRemoteFuel = true
}

func Fuel(ctx context.Context, ft *FuelTank, local bool) (entity.Fuel, error) {
	return fuel(ctx, ft, local)
}

func (vm *MusicGangVM) Meter(infoChan chan<- string, errChan chan<- error) {
	vm.meter(infoChan, errChan)
}

func (vm *MusicGangVM) Cancel() {
	vm.cancel()
}

func (vm *MusicGangVM) MakeOperation(ctx context.Context, ref service.VmCallable, fn VmFunc) (res interface{}, err error) {
	return vm.makeOperation(ctx, ref, fn)
}

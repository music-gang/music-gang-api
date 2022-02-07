package mgvm

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
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

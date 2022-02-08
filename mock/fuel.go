package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.FuelTankService = (*FuelTankService)(nil)

type FuelTankService struct {
	BurnFn func(ctx context.Context, fuel entity.Fuel) error

	FuelFn func(ctx context.Context) (entity.Fuel, error)

	RefuelFn func(ctx context.Context, fuelToRefill entity.Fuel) error

	StatsFn func(ctx context.Context) (*entity.FuelStat, error)
}

func (ft *FuelTankService) Burn(ctx context.Context, fuel entity.Fuel) error {
	if ft.BurnFn == nil {
		panic("BurnFn is not defined")
	}
	return ft.BurnFn(ctx, fuel)
}

func (ft *FuelTankService) Fuel(ctx context.Context) (entity.Fuel, error) {
	if ft.FuelFn == nil {
		panic("FuelFn is not defined")
	}
	return ft.FuelFn(ctx)
}

func (ft *FuelTankService) Refuel(ctx context.Context, fuelToRefill entity.Fuel) error {
	if ft.RefuelFn == nil {
		panic("RefuelFn is not defined")
	}
	return ft.RefuelFn(ctx, fuelToRefill)
}

func (ft *FuelTankService) Stats(ctx context.Context) (*entity.FuelStat, error) {
	if ft.StatsFn == nil {
		panic("StatsFn is not defined")
	}
	return ft.StatsFn(ctx)
}

type FuelTankServiceNoOp struct{}

func (ft *FuelTankServiceNoOp) Burn(ctx context.Context, fuel entity.Fuel) error {
	return nil
}

func (ft *FuelTankServiceNoOp) Fuel(ctx context.Context) (entity.Fuel, error) {
	return entity.Fuel(0), nil
}

func (ft *FuelTankServiceNoOp) Refuel(ctx context.Context, fuelToRefill entity.Fuel) error {
	return nil
}

func (ft *FuelTankServiceNoOp) Stats(ctx context.Context) (*entity.FuelStat, error) {
	return &entity.FuelStat{}, nil
}

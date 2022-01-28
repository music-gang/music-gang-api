package mgvm

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/util"
)

// MusicGangVM is a virtual machine for the Mg language(nodeJS for now).
type MusicGangVM struct {
	ctx      context.Context
	FuelTank *FuelTank
}

// MusicGangVM creates a new MusicGangVM.
// It should be called only once.
func NewMusicGangVM() *MusicGangVM {
	return &MusicGangVM{}
}

func (mg *MusicGangVM) Run(ctx context.Context) error {
	mg.ctx = ctx
	return nil
}

type Action struct {
	ctx             context.Context
	res             <-chan util.Result
	fuelConsumption entity.Fuel
	ContractMG      *entity.Contract
}

type Scheduler struct {
	ctx   context.Context
	queue []*Action
}

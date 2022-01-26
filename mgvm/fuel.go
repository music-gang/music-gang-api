package mgvm

import (
	"sync/atomic"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
)

// Fuel is a virtual unit of measure for the power consuption of the MusicGang VM.
// Like the real world, the fuel is limited, so the VM will stop if the fuel is low.
type Fuel uint64

// Fuel*ActionCost rappresents the cost of an action.
// Greater is the execution time, greater is the cost.
const (
	FuelQuickActionCost   = Fuel(2)
	FuelFastestActionCost = Fuel(3)
	FuelFastActionCost    = Fuel(5)
	FuelMidActionCost     = Fuel(8)
	FuelSlowActionCost    = Fuel(10)
	GasExtremeActionCost  = Fuel(20)
)

var (
	// fuelCostTable is a grid of fuel costs based on the execution time.
	fuelCostTable = map[time.Duration]Fuel{
		time.Millisecond * 200:  FuelQuickActionCost,
		time.Millisecond * 500:  FuelFastestActionCost,
		time.Millisecond * 750:  FuelFastActionCost,
		time.Millisecond * 1200: FuelMidActionCost,
		time.Millisecond * 2000: FuelSlowActionCost,
		time.Millisecond * 3000: GasExtremeActionCost,
	}
)

// FuelCost returns the cost of an action based only on the execution time.
func FuelCost(execution time.Duration) Fuel {
	for k, v := range fuelCostTable {
		if execution < k {
			return v
		}
	}
	return GasExtremeActionCost
}

// ErrFuelTankCapacity is returned when the initial capacity is greater than the max capacity.
var ErrFuelTankCapacity = apperr.Errorf(apperr.EINTERNAL, "fuel tank capacity is less than initial capacity")
var ErrFuelTankFull = apperr.Errorf(apperr.EINTERNAL, "fuel tank is full")
var ErrFuelTankEmpty = apperr.Errorf(apperr.EINTERNAL, "fuel tank is empty")
var ErrFuelTankNotEnough = apperr.Errorf(apperr.EINTERNAL, "fuel tank is not enough")

// FuelTank is the fuel tank of the MusicGang VM.
type FuelTank struct {
	fuel    Fuel
	fuelMax Fuel
}

// NewFuelTank creates a new FuelTank.
func NewFuelTank(maxCap, initialCap Fuel) (*FuelTank, error) {

	if maxCap < initialCap {
		return nil, ErrFuelTankCapacity
	}

	ft := &FuelTank{
		fuel:    initialCap,
		fuelMax: maxCap,
	}

	return ft, nil
}

// Consume consumes the specified amount of fuel.
// It is thread-safe.
func (ft *FuelTank) Consume(fuel Fuel) error {

	if fuel == 0 {
		return ErrFuelTankEmpty
	}

	if ft.Fuel() < fuel {
		return ErrFuelTankNotEnough
	}

	atomic.AddUint64((*uint64)(&ft.fuel), ^uint64(fuel-1))

	return nil
}

// Fuel returns the current amount of fuel.
// It is thread-safe.
func (ft *FuelTank) Fuel() Fuel {
	return Fuel(atomic.LoadUint64((*uint64)(&ft.fuel)))
}

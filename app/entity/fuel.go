package entity

import (
	"time"
)

// Fuel is a virtual unit of measure for the power consuption of the MusicGang VM.
// Like the real world, the fuel is limited, so the VM will stop if the fuel usage reaches the max capacity.
type Fuel uint64

// Fuel*ActionCost rappresents the cost of an action.
// Greater is the execution time, greater is the cost.
const (
	FuelQuickActionAmount   = Fuel(2)
	FuelFastestActionAmount = Fuel(3)
	FuelFastActionAmount    = Fuel(5)
	FuelMidActionAmount     = Fuel(8)
	FuelSlowActionAmount    = Fuel(10)
	GasExtremeActionAmount  = Fuel(20)
)

// vFuel rappresents the virtual units of measure for the power consuption of the MusicGang VM.
const (
	vFuel  = Fuel(1) // virtual fuel unit for the fuel tank
	vKFuel = Fuel(1024)
	vMFuel = vKFuel << 10
	vGFuel = vMFuel << 10
	vTFuel = vGFuel << 10

	FuelTankCapacity = vTFuel
)

var (
	// fuelAmountTable is a grid of fuel costs based on the execution time.
	fuelAmountTable = map[time.Duration]Fuel{
		time.Millisecond * 100:  FuelQuickActionAmount,
		time.Millisecond * 200:  FuelFastestActionAmount,
		time.Millisecond * 450:  FuelFastActionAmount,
		time.Millisecond * 1000: FuelMidActionAmount,
		time.Millisecond * 2000: FuelSlowActionAmount,
		time.Millisecond * 3000: GasExtremeActionAmount,
	}
)

// FuelAmount returns the cost of an action based only on the execution time.
func FuelAmount(execution time.Duration) Fuel {
	for k, v := range fuelAmountTable {
		if execution < k {
			return v
		}
	}
	return GasExtremeActionAmount
}

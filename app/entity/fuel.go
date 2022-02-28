package entity

import (
	"strconv"
	"time"
)

// Fuel is a virtual unit of measure for the power consuption of the MusicGang VM.
// Like the real world, the fuel is limited, so the VM will stop if the fuel usage reaches the max capacity.
type Fuel uint64

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (f Fuel) MarshalBinary() (data []byte, err error) {
	return strconv.AppendUint(nil, uint64(f), 10), nil
}

// Fuel*ActionCost rappresents the cost of an action.
// Greater is the execution time, greater is the cost.
const (
	FuelInstantActionAmount = Fuel(50)
	FuelQuickActionAmount   = Fuel(200)
	FuelFastestActionAmount = Fuel(400)
	FuelFastActionAmount    = Fuel(600)
	FuelMidActionAmount     = Fuel(800)
	FuelSlowActionAmount    = Fuel(1200)
	FuelExtremeActionAmount = Fuel(2500)
	FuelLongActionAmount    = Fuel(5000)
)

// vFuel rappresents the virtual units of measure for the power consuption of the MusicGang VM.
const (
	vFuel  = Fuel(1) // virtual fuel unit for the fuel tank
	vKFuel = Fuel(1024)
	vMFuel = vKFuel << 10
	vGFuel = vMFuel << 10
	vTFuel = vGFuel << 10
)

// FuelRefillAmount retruns how much fuel is refilled at a time.
// It is equivalent to 5% of the capacity of the fuel tank.
const FuelRefillAmount = Fuel(FuelTankCapacity * 5 / 100)

// FuelTankCapacity is the maximum capacity of the fuel tank.
// TODO: this should be a configurable value.
const FuelTankCapacity = 100 * vKFuel

// MaxExecutionTime returns the maximum execution time of an action.
// TODO: this should be a configurable value.
const MaxExecutionTime = 10 * time.Second

// MinExecutionTime returns the minimum execution time of an action.
// In fact, action can be executed in less than the minimum execution time.
const MinExecutionTime = time.Millisecond * 100

var (
	// fuelAmountTable is a grid of fuel costs based on the execution time.
	// It can be read as:
	// from [0 - 100]ms: FuelInstantActionAmount
	// from (100 - 200]ms: FuelQuickActionAmount
	// from (200 - 300]ms: FuelFastestActionAmount
	// from (300 - 500]ms: FuelFastActionAmount
	// from (500 - 1000]ms: FuelMidActionAmount
	// from (1000 - 2000]ms: FuelSlowActionAmount
	// from (2000 - 5000]ms: FuelExtremeActionAmount
	// over 5000ms: FuelExtremeActionAmount
	fuelAmountTable = map[time.Duration]Fuel{
		time.Millisecond * 100:  FuelInstantActionAmount,
		time.Millisecond * 200:  FuelQuickActionAmount,
		time.Millisecond * 300:  FuelFastestActionAmount,
		time.Millisecond * 500:  FuelFastActionAmount,
		time.Millisecond * 1000: FuelMidActionAmount,
		time.Millisecond * 2000: FuelSlowActionAmount,
		time.Millisecond * 5000: FuelExtremeActionAmount,
	}
)

// FuelAmount returns the cost of an action based only on the execution time.
func FuelAmount(execution time.Duration) Fuel {
	for t, fuel := range fuelAmountTable {
		if execution <= t {
			return fuel
		}
	}
	return FuelLongActionAmount
}

// MaxExecutionTime returns the maximum execution time of the contract.
// MaxExecutionTime is based on max fuel compared with fuelAmountTable.
func MaxExecutionTimeFromFuel(fuel Fuel) time.Duration {
	for t, f := range fuelAmountTable {
		if fuel <= f {
			return t
		}
	}

	return MaxExecutionTime
}

// FuelStats represents the statistics of the fuel tank.
type FuelStat struct {
	FuelCapacity    Fuel      `json:"fuel_capacity"`
	FuelUsed        Fuel      `json:"fuel_used"`
	LastRefuelAmout Fuel      `json:"last_refuel_amount"`
	LastRefuelAt    time.Time `json:"last_refuel_at"`
}

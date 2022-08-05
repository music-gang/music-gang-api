package entity

// VmState represents the state of the MusicGangVM.
type VmState int32

const (
	// StateRunning is the state of the MusicGangVM when it is running.
	StateInitializing VmState = iota
	StateRunning
	StatePaused
	StateStopped
)

// String returns a string representation of the VmState.
func (s VmState) String() string {
	switch s {
	case StateInitializing:
		return "initializing"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// VmOperation is a type for the operations of the MusicGang VM.
type VmOperation string

// Defines the operations of the MusicGang VM.
const (
	VmOperationCreateContract  VmOperation = "create-contract"
	VmOperationExecuteContract VmOperation = "execute-contract"
	VmOperationUpdateContract  VmOperation = "update-contract"
	VmOperationDeleteContract  VmOperation = "delete-contract"

	VmOperationMakeContractRevision VmOperation = "make-contract-revision"

	VmOperationCreateUser VmOperation = "create-user"
	VmOperationUpdateUser VmOperation = "update-user"
	VmOperationDeleteUser VmOperation = "delete-user"

	VmOperationAuthenticate VmOperation = "authenticate"
	VmOperationCreateAuth   VmOperation = "create-auth"
	VmOperationDeleteAuth   VmOperation = "delete-auth"

	VmOperationVmStats VmOperation = "vm-stats"

	VmOperationGeneric VmOperation = "generic"
)

const (
	// NotDefinedOperationCost is the default fuel used if not specified by the VmCall when
	NotDefinedOperationCost = Fuel(25)
	// StateFulOperationCost is the extra fuel used when the operation is executed in the full state.
	StateFulOperationCost = Fuel(500)
)

// vmOperationCostTable is a map of operations to their costs.
var vmOperationCostTable = map[VmOperation]Fuel{
	VmOperationExecuteContract:      0, // This 0 because every contract/revision declares its own cost.
	VmOperationCreateContract:       Fuel(10),
	VmOperationUpdateContract:       Fuel(5),
	VmOperationDeleteContract:       Fuel(15),
	VmOperationMakeContractRevision: Fuel(5),
	VmOperationCreateUser:           Fuel(15),
	VmOperationUpdateUser:           Fuel(5),
	VmOperationDeleteUser:           Fuel(10),
	VmOperationAuthenticate:         Fuel(20),
	VmOperationCreateAuth:           Fuel(5),
	VmOperationDeleteAuth:           Fuel(5),
	VmOperationVmStats:              Fuel(0),
}

// VmOperationCost returns the cost of the operation.
func VmOperationCost(op VmOperation) Fuel {
	if cost, ok := vmOperationCostTable[op]; ok {
		return cost
	}
	return NotDefinedOperationCost
}

// IsValidOperation returns true if the operation is valid.
func IsValidOperation(op VmOperation) bool {
	_, ok := vmOperationCostTable[op]
	return ok
}

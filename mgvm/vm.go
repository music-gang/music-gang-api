package mgvm

// MGVM is a virtual machine for the Mg language.
type MGVM struct {
	ft *FuelTank
}

// NewMGVM creates a new MGVM.
// It should be called only once.
func NewMGVM() *MGVM {
	return &MGVM{}
}

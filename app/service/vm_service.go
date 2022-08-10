package service

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
)

const (
	// EngineExecutionTimeoutPanic is the panic message when the engine execution time is exceeded.
	EngineExecutionTimeoutPanic = "engine-execution-panic-timeout"
)

type EngineStateService interface {
	// State returns the state of the engine.
	State() entity.VmState
}

// EngineService is the interface for the engine service.
type EngineService interface {
	ContractExecutorService
	EngineStateService
	// IsRunning returns true if the engine is running.
	IsRunning() bool
	// Pause pauses the engine.
	Pause() error
	// Resume resumes the engine.
	Resume() error
	// Stop stops the engine.
	Stop() error
}

// VmService is a general service for of the MusicGang VM.
type VmService interface {
	EngineService
	VmCallableService
}

// VmCallableService defines all callable services of the MusicGang VM.
type VmCallableService interface {
	AuthManagmentService
	ContractExecutorService
	ContractManagmentService
	FuelStatsService
	UserManagmentService
}

// VmCallable is the interface for the MusicGang VM callable operations.
// Everyone who wants to call the MusicGang VM must implement this interface.
type VmCallable interface {
	// Caller returns the caller of the contract.
	// Can be nil if the Caller is not defined.
	Caller() *entity.User
	// Contract returns the contract that is being called.
	// Can be nil if the Contract is not defined.
	Contract() *entity.Contract
	// Fuel returns only the fuel to perform the choosen operation.
	Fuel() entity.Fuel
	// MaxFuel returns the maximum fuel that the caller can use.
	MaxFuel() entity.Fuel
	// Operation returns the operation that is being called.
	Operation() entity.VmOperation
	// Revision is the revision of the contract that is being called.
	// Can be nil if the Revision is not defined.
	Revision() *entity.Revision
	// WithEngineState returns true if the engine state should not be ignored.
	WithEngineState() bool
	// WithRefuel returns true if is necessary to refuel remaining fuel after Call ends.
	WithRefuel() bool
}

var _ VmCallable = (*VmCall)(nil)

// VmCall rappresents a request from an user to call a contract.
type VmCall struct {
	// ContractRef is the contract that is being called.
	ContractRef *entity.Contract `json:"contract"`

	// User is the user that is calling the vm.
	User *entity.User `json:"caller"`

	// RevisionRef is the revision of the contract that is being called.
	// If RevisionRef is nil but ContractRef is not nil, RevisionRef is set to ContractRef.LastRevision.
	RevisionRef *entity.Revision `json:"revision"`

	// CustomMaxFuel is the maximum fuel that can be used to call the vm.
	CustomMaxFuel *entity.Fuel `json:"custom_max_fuel"`

	// VmOperation is the operation that is being called.
	VmOperation entity.VmOperation `json:"operation"`

	// IgnoreRefuel is true if the remaining fuel must not be refueled.
	IgnoreRefuel bool `json:"ignore_refuel"`

	// IgnoreEngineState ignores the engine state and executes the call anyway.
	IgnoreEngineState bool `json:"ignore_engine_state"`
}

// VmCallOpt is the options for the contract call constructor
type VmCallOpt struct {
	ContractRef       *entity.Contract
	User              *entity.User
	RevisionRef       *entity.Revision
	CustomMaxFuel     *entity.Fuel
	VmOperation       entity.VmOperation
	IgnoreRefuel      bool
	IgnoreEngineState bool
}

// NewVmCall creates a new contract call.
func NewVmCall() *VmCall {
	return NewVmCallWithConfig(VmCallOpt{})
}

// NewVmCallWithConfig creates a new contract call with the given config.
func NewVmCallWithConfig(opt VmCallOpt) *VmCall {
	return &VmCall{
		ContractRef:       opt.ContractRef,
		User:              opt.User,
		RevisionRef:       opt.RevisionRef,
		CustomMaxFuel:     opt.CustomMaxFuel,
		VmOperation:       opt.VmOperation,
		IgnoreRefuel:      opt.IgnoreRefuel,
		IgnoreEngineState: opt.IgnoreEngineState,
	}
}

// Caller returns the caller of the contract.
func (c *VmCall) Caller() *entity.User {
	return c.User
}

// Contract returns the contract that is being called.
func (c VmCall) Contract() *entity.Contract {
	return c.ContractRef
}

// Fuel returns the fuel to perform the choosen operation.
func (c *VmCall) Fuel() entity.Fuel {
	if c.CustomMaxFuel != nil {
		return *c.CustomMaxFuel
	} else if c.RevisionRef != nil {
		return c.RevisionRef.MaxFuel
	} else if c.ContractRef != nil {
		return c.ContractRef.MaxFuel
	} else {
		return entity.VmOperationCost(c.Operation())
	}
}

// MaxFuel returns the maximum fuel that can be used to call the contract.
func (c *VmCall) MaxFuel() entity.Fuel {

	stateFulFuel := entity.Fuel(0)
	if c.ContractRef != nil && c.ContractRef.Stateful {
		stateFulFuel = entity.StateFulOperationCost
	}

	return c.Fuel() + stateFulFuel
}

// Operation returns the operation type that is being called.
func (c *VmCall) Operation() entity.VmOperation {
	if c.VmOperation == "" {
		return entity.VmOperationGeneric
	}
	return c.VmOperation
}

// Revision is the revision of the contract that is being called.
func (c *VmCall) Revision() *entity.Revision {
	if c.RevisionRef != nil {
		return c.RevisionRef
	} else if c.ContractRef != nil && c.ContractRef.LastRevision != nil {
		return c.ContractRef.LastRevision
	} else {
		return nil
	}
}

// WithEngineState returns true if the engine state should not be ignored.
func (c *VmCall) WithEngineState() bool {
	return !c.IgnoreEngineState
}

// WithRefuel returns true if is necessary to refuel remaining fuel after Call ends.
func (c *VmCall) WithRefuel() bool {
	return !c.IgnoreRefuel
}

// CPUsPoolService is the interface for the CPUsPool service.
// It is used from MusicGang VM to acquire and release cores in order to execute operations inside the VM.
type CPUsPoolService interface {
	// AcquireCore acquires a core from the CPUsPool.
	// Returns a function that must be called when the core is not needed anymore.
	// CORE LEAK IF NOT CALLED.
	AcquireCore(ctx context.Context, call VmCallable) (release func(), err error)
}

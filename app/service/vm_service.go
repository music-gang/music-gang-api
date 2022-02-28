package service

import (
	"github.com/music-gang/music-gang-api/app/entity"
)

const (
	// EngineExecutionTimeoutPanic is the panic message when the engine execution time is exceeded.
	EngineExecutionTimeoutPanic = "engine-execution-panic-timeout"
)

// EngineService is the interface for the engine service.
type EngineService interface {
	ContractExecutorService
	// IsRunning returns true if the engine is running.
	IsRunning() bool
	// Pause pauses the engine.
	Pause() error
	// Resume resumes the engine.
	Resume() error
	// State returns the state of the engine.
	State() entity.State
	// Stop stops the engine.
	Stop() error
}

// VmService is a service for the engine of the MusicGang VM.
type VmService interface {
	FuelMeterService
	EngineService
}

// VmCallableService defines all callable services of the MusicGang VM.
type VmCallableService interface {
	FuelMeterService
	ContractManagmentService
	ContractExecutorService
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
	// MaxFuel returns the maximum fuel that the caller can use.
	MaxFuel() entity.Fuel
	// Operation returns the operation that is being called.
	Operation() entity.VmOperation
	// Revision is the revision of the contract that is being called.
	// Can be nil if the Revision is not defined.
	Revision() *entity.Revision
}

var _ VmCallable = (*VmCall)(nil)

// VmCall rappresents a request from an user to call a contract.
type VmCall struct {
	ContractRef   *entity.Contract   `json:"contract"`
	User          *entity.User       `json:"caller"`
	RevisionRef   *entity.Revision   `json:"revision"`
	CustomMaxFuel *entity.Fuel       `json:"custom_max_fuel"`
	VmOperation   entity.VmOperation `json:"operation"`
}

// VmCallOpt is the options for the contract call constructor
type VmCallOpt struct {
	ContractRef   *entity.Contract
	User          *entity.User
	RevisionRef   *entity.Revision
	CustomMaxFuel *entity.Fuel
	VmOperation   entity.VmOperation
}

// NewVmCall creates a new contract call.
func NewVmCall() *VmCall {
	return NewVmCallWithConfig(VmCallOpt{})
}

// NewVmCallWithConfig creates a new contract call with the given config.
func NewVmCallWithConfig(opt VmCallOpt) *VmCall {
	return &VmCall{
		ContractRef:   opt.ContractRef,
		User:          opt.User,
		RevisionRef:   opt.RevisionRef,
		CustomMaxFuel: opt.CustomMaxFuel,
		VmOperation:   opt.VmOperation,
	}
}

// Caller returns the caller of the contract.
func (c VmCall) Caller() *entity.User {
	return c.User
}

// Contract returns the contract that is being called.
func (c VmCall) Contract() *entity.Contract {
	return c.ContractRef
}

// MaxFuel returns the maximum fuel that can be used to call the contract.
func (c *VmCall) MaxFuel() entity.Fuel {
	if c.CustomMaxFuel != nil {
		return *c.CustomMaxFuel
	} else if c.RevisionRef != nil {
		return c.RevisionRef.MaxFuel
	} else if c.ContractRef != nil {
		return c.ContractRef.MaxFuel
	} else {
		return entity.NotDefinedOperationCost
	}
}

// Operation returns the operation type that is being called.
func (c *VmCall) Operation() entity.VmOperation {
	return c.VmOperation
}

// Revision is the revision of the contract that is being called.
func (c VmCall) Revision() *entity.Revision {
	if c.RevisionRef != nil {
		return c.RevisionRef
	} else if c.ContractRef != nil && c.ContractRef.LastRevision != nil {
		return c.ContractRef.LastRevision
	} else {
		return nil
	}
}

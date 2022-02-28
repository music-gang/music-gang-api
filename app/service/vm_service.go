package service

import (
	"github.com/music-gang/music-gang-api/app/entity"
)

// State represents the state of the MusicGangVM.
type State int32

const (
	// StateRunning is the state of the MusicGangVM when it is running.
	StateInitializing State = iota
	StateRunning
	StatePaused
	StateStopped
)

const (
	// EngineExecutionTimeoutPanic is the panic message when the engine execution time is exceeded.
	EngineExecutionTimeoutPanic = "engine-execution-panic-timeout"
)

// String returns a string representation of the State.
func (s State) String() string {
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
	State() State
	// Stop stops the engine.
	Stop() error
}

// VmService is a service for the engine of the MusicGang VM.
type VmService interface {
	FuelMeterService
	EngineService
}

// VmCaller is the interface for the MusicGang VM callable operations.
// Everyone who wants to call the MusicGang VM must implement this interface.
type VmCaller interface {
	MaxFuel() entity.Fuel
}

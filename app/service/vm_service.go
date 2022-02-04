package service

import (
	"context"
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

// VmService is a service for the engine of the MusicGang VM.
type VmService interface {
	// ExecContract executes a contract.
	ExecContract(ctx context.Context, contractRef *ContractCall) (res interface{}, err error)
	// IsRunning returns true if the engine is running.
	IsRunning() bool
	// Pause pauses the engine.
	Pause()
	// Resume resumes the engine.
	Resume()
	// State returns the state of the engine.
	State() State
	// Stop stops the engine.
	Stop()
}

package mgvm

import (
	"context"
	"sync/atomic"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.EngineService = (*Engine)(nil)

// Engine is the state machine for the MusicGangVM.
// His job is to exec the called contracts.
type Engine struct {
	state     entity.State
	Executors map[entity.RevisionVersion]service.ContractExecutorService
}

// NewEngine creates a new engine.
func NewEngine() *Engine {
	return &Engine{
		state:     entity.StateInitializing,
		Executors: make(map[entity.RevisionVersion]service.ContractExecutorService),
	}
}

// ExecContract effectively executes the contract and returns the result.
// If the engine goes into execution timeout, it panics with EngineExecutionTimeoutPanic.
func (e *Engine) ExecContract(ctx context.Context, revision *entity.Revision) (res interface{}, err error) {

	if !e.IsRunning() {
		return nil, apperr.Errorf(apperr.EMGVM, "engine is not running")
	}

	executor, err := e.getExecutor(revision.Version)
	if err != nil {
		return nil, err
	}

	return executor.ExecContract(ctx, revision)
}

// IsRunning returns true if the engine is running.
func (e *Engine) IsRunning() bool {
	return atomic.LoadInt32((*int32)(&e.state)) == int32(entity.StateRunning)
}

// Pause pauses the engine.
func (e *Engine) Pause() error {
	// It is not possible to pause the engine if is stopped or already paused.
	if e.State() == entity.StateStopped || e.State() == entity.StatePaused {
		return apperr.Errorf(apperr.EMGVM, "engine is not running")
	}
	atomic.StoreInt32((*int32)(&e.state), int32(entity.StatePaused))
	return nil
}

// Resume resumes the engine.
func (e *Engine) Resume() error {
	// it is not possible to resume the engine if is already running.
	if e.IsRunning() {
		return apperr.Errorf(apperr.EMGVM, "engine is already running")
	}
	atomic.StoreInt32((*int32)(&e.state), int32(entity.StateRunning))
	return nil
}

// State returns the state of the engine.
func (e *Engine) State() entity.State {
	return entity.State(atomic.LoadInt32((*int32)(&e.state)))
}

// Stop stops the engine.
func (e *Engine) Stop() error {
	// It is not possible to stop the engine if is already stopped.
	if e.State() == entity.StateStopped {
		return apperr.Errorf(apperr.EMGVM, "engine is already stopped")
	}
	atomic.StoreInt32((*int32)(&e.state), int32(entity.StateStopped))
	return nil
}

// getExecutor returns the executor for the given revision version.
func (e *Engine) getExecutor(version entity.RevisionVersion) (service.ContractExecutorService, error) {
	ex, ok := e.Executors[version]
	if !ok {
		return nil, apperr.Errorf(apperr.EMGVM, "no executor for revision version '%s'", version)
	}
	return ex, nil
}

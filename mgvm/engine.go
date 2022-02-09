package mgvm

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/robertkrimen/otto"
)

var _ service.EngineService = (*Engine)(nil)

const (
	// EngineExecutionTimeoutPanic is the panic message when the engine execution time is exceeded.
	EngineExecutionTimeoutPanic = "engine-execution-panic-timeout"
)

// Engine is the state machine for the MusicGangVM.
// His job is to exec the called contracts.
type Engine struct {
	state service.State
}

// NewEngine creates a new engine.
func NewEngine() *Engine {
	return &Engine{
		state: service.StateInitializing,
	}
}

// ExecContract effectively executes the contract and returns the result.
// If the engine goes into execution timeout, it panics with EngineExecutionTimeoutPanic.
func (e *Engine) ExecContract(ctx context.Context, contractRef *service.ContractCall) (res interface{}, err error) {

	if !e.IsRunning() {
		return nil, apperr.Errorf(apperr.EMGVM, "engine is not running")
	}

	select {
	case <-ctx.Done():
		return nil, apperr.Errorf(apperr.EMGVM, "Timeout while executing contract")
	default:
	}

	ottoVm := otto.New()
	ottoVm.Interrupt = make(chan func(), 1)

	timeoutTicker := time.NewTicker(contractRef.Contract.MaxExecutionTime())
	defer timeoutTicker.Stop()

	go func() {
		<-timeoutTicker.C
		ottoVm.Interrupt <- func() {
			panic(EngineExecutionTimeoutPanic)
		}
	}()

	_, err = ottoVm.Run(contractRef.Contract.LastRevision.Code)
	close(ottoVm.Interrupt)

	if err != nil {
		return nil, apperr.Errorf(apperr.EMGVM, "Error while executing contract: %s", err.Error())
	}

	value, err := ottoVm.Get("result")
	if err != nil {
		return nil, apperr.Errorf(apperr.EMGVM, "Error while retrieving result: %s", err.Error())
	}

	str, err := value.ToString()
	if err != nil {
		return nil, apperr.Errorf(apperr.EMGVM, "Error while parsing contract result: %s", err.Error())
	}

	return str, nil
}

// IsRunning returns true if the engine is running.
func (e *Engine) IsRunning() bool {
	return atomic.LoadInt32((*int32)(&e.state)) == int32(service.StateRunning)
}

// Pause pauses the engine.
func (e *Engine) Pause() error {
	// It is not possible to pause the engine if is stopped or already paused.
	if e.State() == service.StateStopped || e.State() == service.StatePaused {
		return apperr.Errorf(apperr.EMGVM, "engine is not running")
	}
	atomic.StoreInt32((*int32)(&e.state), int32(service.StatePaused))
	return nil
}

// Resume resumes the engine.
func (e *Engine) Resume() error {
	// it is not possible to resume the engine if is already running.
	if e.IsRunning() {
		return apperr.Errorf(apperr.EMGVM, "engine is already running")
	}
	atomic.StoreInt32((*int32)(&e.state), int32(service.StateRunning))
	return nil
}

// State returns the state of the engine.
func (e *Engine) State() service.State {
	return service.State(atomic.LoadInt32((*int32)(&e.state)))
}

// Stop stops the engine.
func (e *Engine) Stop() error {
	// It is not possible to stop the engine if is already stopped.
	if e.State() == service.StateStopped {
		return apperr.Errorf(apperr.EMGVM, "engine is already stopped")
	}
	atomic.StoreInt32((*int32)(&e.state), int32(service.StateStopped))
	return nil
}

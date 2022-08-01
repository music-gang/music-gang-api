package executor

import (
	"context"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/robertkrimen/otto"
)

var _ service.ContractExecutorService = (*AnchorageContractExecutor)(nil)

// AnchorageContractExecutor is the contract executor for the anchorage contract version.
type AnchorageContractExecutor struct{}

// NewAnchorageContractExecutor creates a new AnchorageContractExecutor.
func NewAnchorageContractExecutor() *AnchorageContractExecutor {
	return &AnchorageContractExecutor{}
}

// ExecContract effectively executes the contract and returns the result.
// If the engine goes into execution timeout, it panics with EngineExecutionTimeoutPanic.
func (*AnchorageContractExecutor) ExecContract(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {

	revision, err := opt.Revision()
	if err != nil {
		return nil, err
	}

	contract, err := opt.Contract()
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, apperr.Errorf(apperr.EANCHORAGE, "Timeout while executing contract")
	default:
	}

	ottoVm := otto.New()
	ottoVm.Interrupt = make(chan func(), 1)

	timeoutTicker := time.NewTicker(entity.MaxExecutionTimeFromFuel(revision.MaxFuel))
	defer timeoutTicker.Stop()

	go func() {
		<-timeoutTicker.C
		ottoVm.Interrupt <- func() {
			panic(service.EngineExecutionTimeoutPanic)
		}
	}()

	if contract.Stateful && opt.StateRef != nil && opt.StateRef.Value != nil {
		injectStateAccessor(ottoVm, opt.StateRef)
	}

	_, err = ottoVm.Run(revision.CompiledCode)
	close(ottoVm.Interrupt)

	if err != nil {
		return nil, apperr.Errorf(apperr.EANCHORAGE, "Error while executing contract: %s", err.Error())
	}

	value, err := ottoVm.Get("result")
	if err != nil {
		return nil, apperr.Errorf(apperr.EANCHORAGE, "Error while retrieving result: %s", err.Error())
	}

	str, err := value.ToString()
	if err != nil {
		return nil, apperr.Errorf(apperr.EANCHORAGE, "Error while parsing contract result: %s", err.Error())
	}

	return str, nil
}

// injectStateAccessor injects the state accessor into the otto vm.
func injectStateAccessor(vm *otto.Otto, contractState *entity.State) {
	vm.Set("setState", func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 2 {
			return otto.UndefinedValue()
		}
		key, err := call.Argument(0).ToString()
		if err != nil {
			return otto.UndefinedValue()
		}
		var value any

		value, err = call.Argument(1).Export()
		if err != nil {
			return otto.UndefinedValue()
		}

		contractState.Value[key] = value
		return otto.UndefinedValue()
	})

	vm.Set("getState", func(call otto.FunctionCall) otto.Value {

		if len(call.ArgumentList) != 1 {
			return otto.UndefinedValue()
		}

		key, err := call.Argument(0).ToString()
		if err != nil {
			return otto.UndefinedValue()
		}

		value, ok := contractState.Value[key]
		if !ok {
			return otto.UndefinedValue()
		}

		ottoValue, err := otto.ToValue(value)
		if err != nil {
			return otto.UndefinedValue()
		}

		return ottoValue
	})
}

package executor

import (
	"context"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/robertkrimen/otto"
)

// AnchorageContractExecutor is the contract executor for the anchorage contract version.
type AnchorageContractExecutor struct{}

// NewAnchorageContractExecutor creates a new AnchorageContractExecutor.
func NewAnchorageContractExecutor() *AnchorageContractExecutor {
	return &AnchorageContractExecutor{}
}

// ExecContract effectively executes the contract and returns the result.
// If the engine goes into execution timeout, it panics with EngineExecutionTimeoutPanic.
func (*AnchorageContractExecutor) ExecContract(ctx context.Context, contractRef *service.ContractCall) (res interface{}, err error) {

	select {
	case <-ctx.Done():
		return nil, apperr.Errorf(apperr.EANCHORAGE, "Timeout while executing contract")
	default:
	}

	ottoVm := otto.New()
	ottoVm.Interrupt = make(chan func(), 1)

	timeoutTicker := time.NewTicker(contractRef.Contract.MaxExecutionTime())
	defer timeoutTicker.Stop()

	go func() {
		<-timeoutTicker.C
		ottoVm.Interrupt <- func() {
			panic(service.EngineExecutionTimeoutPanic)
		}
	}()

	_, err = ottoVm.Run(contractRef.Contract.LastRevision.Code)
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

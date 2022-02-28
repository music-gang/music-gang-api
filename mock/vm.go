package mock

// import (
// 	"context"

// 	"github.com/music-gang/music-gang-api/app/entity"
// 	"github.com/music-gang/music-gang-api/app/service"
// )

// var _ service.VmContractExecCall = (*emptyVmContractExecCall)(nil)

// type emptyVmContractExecCall struct{}

// func (e *emptyVmContractExecCall) Context() context.Context {
// 	return context.TODO()
// }

// func (e *emptyVmContractExecCall) MaxFuel() entity.Fuel {
// 	return entity.Fuel(0)
// }

// func (e *emptyVmContractExecCall) Operation() entity.VmOperation {
// 	return entity.VmOperation(rune(0))
// }

// func (e *emptyVmContractExecCall) Caller() *entity.User {
// 	return nil
// }

// func (e *emptyVmContractExecCall) Contract() *entity.Contract {
// 	return nil
// }

// func (e *emptyVmContractExecCall) Revision() *entity.Revision {
// 	return nil
// }

// func MakeEmptyVmContractExecCall() service.VmContractExecCall {
// 	return &emptyVmContractExecCall{}
// }

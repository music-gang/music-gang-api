package mgvm

import (
	"context"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

// Auhenticate authenticates a user.
// This func is a wrapper for the AuthManagmentService.Auhenticate.
// No checks is performed.
func (vm *MusicGangVM) Auhenticate(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {

	opMaxFuel := entity.VmOperationCost(entity.VmOperationAuthenticate)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          app.UserFromContext(ctx),
		CustomMaxFuel: &opMaxFuel,
		IgnoreRefuel:  true,
		VmOperation:   entity.VmOperationAuthenticate,
	})

	res, err := vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return vm.AuthManagmentService.Auhenticate(ctx, opts)
	})
	if err != nil {
		return nil, err
	}

	if res, ok := res.(*entity.Auth); !ok || res == nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "invalid auth result")
	}

	return res.(*entity.Auth), nil
}

// CreateAuth creates a new auth.
// This call consumes fuel.
// No check on authorization is performed.
func (vm *MusicGangVM) CreateAuth(ctx context.Context, auth *entity.Auth) error {

	user := app.UserFromContext(ctx)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          user,
		CustomMaxFuel: nil,
		IgnoreRefuel:  true,
		VmOperation:   entity.VmOperationCreateAuth,
	})

	_, err := vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return nil, vm.AuthManagmentService.CreateAuth(ctx, auth)
	})

	return err
}

// CreateContract creates a new contract under a vm operation.
// This call consumes fuel.
// No check on authorization is performed.
func (vm *MusicGangVM) CreateContract(ctx context.Context, contract *entity.Contract) (err error) {
	user := app.UserFromContext(ctx)

	opMaxFuel := entity.VmOperationCost(entity.VmOperationCreateContract)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          user,
		CustomMaxFuel: &opMaxFuel,
		IgnoreRefuel:  true,
		ContractRef:   contract,
		VmOperation:   entity.VmOperationCreateContract,
	})

	_, err = vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return nil, vm.ContractManagmentService.CreateContract(ctx, ref.Contract())
	})
	return err
}

// CreateUser creates a new user.
// This call consumes fuel.
func (vm *MusicGangVM) CreateUser(ctx context.Context, user *entity.User) error {

	opMaxFuel := entity.VmOperationCost(entity.VmOperationCreateUser)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          user,
		IgnoreRefuel:  true,
		CustomMaxFuel: &opMaxFuel,
		VmOperation:   entity.VmOperationCreateUser,
	})

	_, err := vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return nil, vm.UserManagmentService.CreateUser(ctx, ref.Caller())
	})

	return err
}

// DeleteAuth deletes the auth.
// This call consumes fuel.
// No check on authorization is performed.
func (vm *MusicGangVM) DeleteAuth(ctx context.Context, id int64) error {

	user := app.UserFromContext(ctx)

	opMaxFuel := entity.VmOperationCost(entity.VmOperationDeleteAuth)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          user,
		CustomMaxFuel: &opMaxFuel,
		IgnoreRefuel:  true,
		VmOperation:   entity.VmOperationDeleteAuth,
	})

	_, err := vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return nil, vm.AuthManagmentService.DeleteAuth(ctx, id)
	})

	return err
}

// DeleteContract deletes the contract under a vm operation.
// This call consumes fuel.
// No check on authorization is performed.
func (vm *MusicGangVM) DeleteContract(ctx context.Context, id int64) error {

	user := app.UserFromContext(ctx)

	opMaxFuel := entity.VmOperationCost(entity.VmOperationDeleteContract)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          user,
		CustomMaxFuel: &opMaxFuel,
		IgnoreRefuel:  true,
		VmOperation:   entity.VmOperationDeleteContract,
	})

	_, err := vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return nil, vm.ContractManagmentService.DeleteContract(ctx, id)
	})

	return err
}

// DeleteUser deletes the user.
// This call consumes fuel.
// No check on authorization is performed.
func (vm *MusicGangVM) DeleteUser(ctx context.Context, id int64) error {

	user := app.UserFromContext(ctx)

	opMaxFuel := entity.VmOperationCost(entity.VmOperationDeleteUser)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          user,
		CustomMaxFuel: &opMaxFuel,
		IgnoreRefuel:  true,
		VmOperation:   entity.VmOperationDeleteUser,
	})

	_, err := vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return nil, vm.UserManagmentService.DeleteUser(ctx, id)
	})

	return err
}

// ExecContract executes the contract.
// This func is a wrapper for the Engine.ExecContract.
func (vm *MusicGangVM) ExecContract(ctx context.Context, opt service.ContractCallOpt) (res interface{}, err error) {

	user := app.UserFromContext(ctx)

	revision, err := opt.Revision()
	if err != nil {
		return nil, err
	}

	contract, err := opt.Contract()
	if err != nil {
		return nil, err
	}

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:        user,
		RevisionRef: revision,
		VmOperation: entity.VmOperationExecuteContract,
		ContractRef: contract,
	})

	return vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {

		if ref.Contract().Stateful {

			needToCreateZeroState := false

			state, err := vm.StateService.FindStateByRevisionID(ctx, ref.Revision().ID)
			if err != nil {
				if errCode := apperr.ErrorCode(err); errCode != apperr.ENOTFOUND {
					return nil, err
				}
				needToCreateZeroState = true
			}

			if needToCreateZeroState {

				state = &entity.State{
					RevisionID: ref.Revision().ID,
					Value:      make(entity.StateValue),
				}

				if err := vm.StateService.CreateState(ctx, state); err != nil {
					return nil, err
				}
			}

			opt.StateRef = state
		}

		res, err := vm.EngineService.ExecContract(ctx, opt)
		if err != nil {
			return nil, err
		}

		if ref.Contract().Stateful {
			if _, err := vm.StateService.UpdateState(ctx, ref.Revision().ID, opt.StateRef.Value); err != nil {
				return nil, err
			}
		}

		return res, nil
	})
}

// MakeRevision makes a revision under a vm operation.
// This call consumes fuel.
// No check on authorization is performed.
func (vm *MusicGangVM) MakeRevision(ctx context.Context, revision *entity.Revision) (err error) {

	user := app.UserFromContext(ctx)

	opMaxFuel := entity.VmOperationCost(entity.VmOperationMakeContractRevision)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          user,
		CustomMaxFuel: &opMaxFuel,
		IgnoreRefuel:  true,
		VmOperation:   entity.VmOperationMakeContractRevision,
		RevisionRef:   revision,
	})

	_, err = vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return nil, vm.ContractManagmentService.MakeRevision(ctx, ref.Revision())
	})

	return err
}

// Stats returns the stats of fuel tank usage.
func (vm *MusicGangVM) Stats(ctx context.Context) (*entity.FuelStat, error) {

	user := app.UserFromContext(ctx)

	opMaxFuel := entity.VmOperationCost(entity.VmOperationVmStats)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:              user,
		CustomMaxFuel:     &opMaxFuel,
		IgnoreRefuel:      true,
		VmOperation:       entity.VmOperationVmStats,
		IgnoreEngineState: true,
	})

	res, err := vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return vm.FuelTank.Stats(ctx)
	})
	if err != nil {
		return nil, err
	}

	if v, ok := res.(*entity.FuelStat); !ok || v == nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "invalid vm stats result")
	}

	return res.(*entity.FuelStat), nil
}

// UpdateContract updates the contract under a vm operation.
// This call consumes fuel.
// No check on authorization is performed.
func (vm *MusicGangVM) UpdateContract(ctx context.Context, id int64, contract service.ContractUpdate) (*entity.Contract, error) {
	user := app.UserFromContext(ctx)

	opMaxFuel := entity.VmOperationCost(entity.VmOperationUpdateContract)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          user,
		CustomMaxFuel: &opMaxFuel,
		IgnoreRefuel:  true,
		VmOperation:   entity.VmOperationUpdateContract,
	})

	result, err := vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return vm.ContractManagmentService.UpdateContract(ctx, id, contract)
	})

	if err != nil {
		return nil, err
	}

	if v, ok := result.(*entity.Contract); !ok || v == nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "invalid contract update result")
	}

	return result.(*entity.Contract), nil
}

// UpdateUser updates the user.
// This call consumes fuel.
// No check on authorization is performed.
func (vm *MusicGangVM) UpdateUser(ctx context.Context, id int64, upd service.UserUpdate) (*entity.User, error) {

	user := app.UserFromContext(ctx)

	opMaxFuel := entity.VmOperationCost(entity.VmOperationUpdateUser)

	call := service.NewVmCallWithConfig(service.VmCallOpt{
		User:          user,
		CustomMaxFuel: &opMaxFuel,
		IgnoreRefuel:  true,
		VmOperation:   entity.VmOperationUpdateUser,
	})

	result, err := vm.makeOperation(ctx, call, func(ctx context.Context, ref service.VmCallable) (interface{}, error) {
		return vm.UserManagmentService.UpdateUser(ctx, id, upd)
	})
	if err != nil {
		return nil, err
	}

	if v, ok := result.(*entity.User); !ok || v == nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "invalid user update result")
	}

	return result.(*entity.User), nil
}

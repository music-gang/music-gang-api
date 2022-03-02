package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.ContractService = (*ContractService)(nil)

type ContractService struct {
	FindContractByIDFn             func(ctx context.Context, id int64) (*entity.Contract, error)
	FindContractsFn                func(ctx context.Context, filter service.ContractFilter) (entity.Contracts, int, error)
	FindRevisionByContractAndRevFn func(ctx context.Context, contractID int64, rev entity.RevisionNumber) (*entity.Revision, error)
	CreateContractFn               func(ctx context.Context, contract *entity.Contract) error
	DeleteContractFn               func(ctx context.Context, id int64) error
	MakeRevisionFn                 func(ctx context.Context, revision *entity.Revision) error
	UpdateContractFn               func(ctx context.Context, id int64, contract service.ContractUpdate) (*entity.Contract, error)
}

func (c *ContractService) FindContractByID(ctx context.Context, id int64) (*entity.Contract, error) {
	if c.FindContractByIDFn == nil {
		panic("FindContractByIDFn is not defined")
	}
	return c.FindContractByIDFn(ctx, id)
}

func (c *ContractService) FindContracts(ctx context.Context, filter service.ContractFilter) (entity.Contracts, int, error) {
	if c.FindContractsFn == nil {
		panic("FindContractsFn is not defined")
	}
	return c.FindContractsFn(ctx, filter)
}

func (c *ContractService) FindRevisionByContractAndRev(ctx context.Context, contractID int64, rev entity.RevisionNumber) (*entity.Revision, error) {
	if c.FindRevisionByContractAndRevFn == nil {
		panic("FindRevisionByContractAndRevFn is not defined")
	}
	return c.FindRevisionByContractAndRevFn(ctx, contractID, rev)
}

func (c *ContractService) CreateContract(ctx context.Context, contract *entity.Contract) error {
	if c.CreateContractFn == nil {
		panic("CreateContractFn is not defined")
	}
	return c.CreateContractFn(ctx, contract)
}

func (c *ContractService) DeleteContract(ctx context.Context, id int64) error {
	if c.DeleteContractFn == nil {
		panic("DeleteContractFn is not defined")
	}
	return c.DeleteContractFn(ctx, id)
}

func (c *ContractService) MakeRevision(ctx context.Context, revision *entity.Revision) error {
	if c.MakeRevisionFn == nil {
		panic("MakeRevisionFn is not defined")
	}
	return c.MakeRevisionFn(ctx, revision)
}

func (c *ContractService) UpdateContract(ctx context.Context, id int64, contract service.ContractUpdate) (*entity.Contract, error) {
	if c.UpdateContractFn == nil {
		panic("UpdateContractFn is not defined")
	}

	return c.UpdateContractFn(ctx, id, contract)
}

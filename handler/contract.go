package handler

import (
	"context"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

// CallContract handles the contract execution business logic.
func (s *ServiceHandler) CallContract(ctx context.Context, contractID int64, revisionNumber entity.RevisionNumber) (res any, err error) {

	var revision *entity.Revision

	if revisionNumber != 0 {
		revision, err = s.ContractSearchService.FindRevisionByContractAndRev(ctx, contractID, revisionNumber)
		if err != nil {
			s.Logger.Error(apperr.ErrorLog(err))
			return nil, err
		}
	} else {

		contract, err := s.ContractSearchService.FindContractByID(ctx, contractID)
		if err != nil {
			s.Logger.Error(apperr.ErrorLog(err))
			return nil, err
		}

		revision = contract.LastRevision
	}

	result, err := s.VmCallableService.ExecContract(ctx, service.ContractCallOpt{
		ContractRef: revision.Contract,
		RevisionRef: revision,
	})
	if err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, err
	}

	return result, nil
}

// CreateContract handles the contract create business logic.
func (s *ServiceHandler) CreateContract(ctx context.Context, contract *entity.Contract) (*entity.Contract, error) {
	if err := s.VmCallableService.CreateContract(ctx, contract); err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, err
	}
	return contract, nil
}

// FindContractByID handles the contract search business logic.
func (s *ServiceHandler) FindContractByID(ctx context.Context, contractID int64) (res *entity.Contract, err error) {
	if contract, err := s.ContractSearchService.FindContractByID(ctx, contractID); err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, err
	} else {
		return contract, nil
	}
}

// MakeContractRevision handles the creation of new revision of a contract business logic.
func (s *ServiceHandler) MakeContractRevision(ctx context.Context, revision *entity.Revision) (*entity.Revision, error) {
	if err := s.VmCallableService.MakeRevision(ctx, revision); err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, err
	}
	return revision, nil
}

// UpdateContract handles the contract update business logic.
func (s *ServiceHandler) UpdateContract(ctx context.Context, contractID int64, params service.ContractUpdate) (*entity.Contract, error) {

	contract, err := s.VmCallableService.UpdateContract(ctx, contractID, params)
	if err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, err
	}

	return contract, nil
}

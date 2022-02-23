package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.ContractService = (*ContractService)(nil)

// ContractService is the postgres implementation of the contract service.
type ContractService struct {
	db *DB
}

// NewContractService creates a new contract service.
func NewContractService(db *DB) *ContractService {
	return &ContractService{db: db}
}

// CreateContract creates a new contract.
// Return EINVALID if the contract is invalid.
// Return EEXISTS if the contract already exists.
// Return EFORBIDDEN if the user is not allowed to create a contract.
// Return EUNAUTHORIZED if the contract owner is not the authenticated user or user is not authenticated.
func (cs *ContractService) CreateContract(ctx context.Context, contract *entity.Contract) error {

	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createContract(ctx, tx, contract); err != nil {
		return err
	} else if err := attachContractAssociations(ctx, tx, contract); err != nil {
		return err
	} else if err := tx.Commit(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return nil
}

// DeleteContract deletes the contract with the given id.
// Return EUNAUTHORIZED if the contract is not the same as the authenticated user.
// Return ENOTFOUND if the contract does not exist.
// This service also deletes the revisions of the contract.
func (cs *ContractService) DeleteContract(ctx context.Context, id int64) error {

	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteContract(ctx, tx, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return nil
}

// FindContractByID returns the contract with the given id.
// Return ENOTFOUND if the contract does not exist.
func (cs *ContractService) FindContractByID(ctx context.Context, id int64) (*entity.Contract, error) {

	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	contract, err := findContractByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := attachContractAssociations(ctx, tx, contract); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return contract, nil
}

// FindContracts returns a list of contracts filtered by the given options.
// Also returns the total count of contracts.
func (cs *ContractService) FindContracts(ctx context.Context, filter service.ContractFilter) (entity.Contracts, int, error) {

	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	return findContracts(ctx, tx, filter)
}

// FindRevisionByContractAndRev returns the revision searched by the given contract and revision number.
// Return ENOTFOUND if the revision does not exist.
// func (cs *ContractService) FindRevisionByContractAndRev(ctx context.Context, contractID int64, rev entity.RevisionNumber) (*entity.Revision, error)

// // FindRevisions returns a list of revisions of the contract with the given id.
// // Also returns the total count of revisions.
// func (cs *ContractService) FindRevisions(ctx context.Context, filter service.RevisionFilter) (entity.Revisions, int, error)

// MakeRevision creates a new revision of the contract.
// Return ENOTFOUND if the contract does not exist.
// Return EINVALID if the revision is invalid.
// It shouldn't return ECONFLICT because there's a UNIQUE constraint on the revision number and the Contract ID.
func (cs *ContractService) MakeRevision(ctx context.Context, revision *entity.Revision) error {

	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if revision.ContractID == 0 {
		return apperr.Errorf(apperr.EINVALID, "contract id is required")
	}

	contract, err := cs.FindContractByID(ctx, revision.ContractID)
	if err != nil {
		return err
	} else if contract.UserID != app.UserIDFromContext(ctx) {
		return apperr.Errorf(apperr.EUNAUTHORIZED, "contract is not owned by the authenticated user")
	}

	var newRevNumber uint = 1
	if contract.LastRevision != nil {
		newRevNumber = uint(contract.LastRevision.Rev) + 1
	}

	if revision.MaxFuel == 0 {
		revision.MaxFuel = contract.MaxFuel
	}
	revision.Rev = entity.RevisionNumber(newRevNumber)

	if err := makeRevision(ctx, tx, revision); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return nil
}

// UpdateContract updates the given contract.
// Return ENOTFOUND if the contract does not exist.
// Return EUNAUTHORIZED if the contract is not owned by the authenticated user.
func (cs *ContractService) UpdateContract(ctx context.Context, id int64, upd service.ContractUpdate) (*entity.Contract, error) {

	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	contract, err := updateContract(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	} else if err := attachContractAssociations(ctx, tx, contract); err != nil {
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to commit transaction: %v", err)
	}

	return contract, nil
}

// attachContractAssociations attaches all associations of the contract to the database.
func attachContractAssociations(ctx context.Context, tx *Tx, contract *entity.Contract) (err error) {
	if contract.User, err = findUserByID(ctx, tx, contract.UserID); err != nil {
		return err
	}
	return nil
}

// createContract takes a contract, validates it, check if user of context is authorized to create the contract, and inserts it into the database.
func createContract(ctx context.Context, tx *Tx, contract *entity.Contract) error {

	contract.CreatedAt = tx.now
	contract.UpdatedAt = contract.CreatedAt

	if err := contract.Validate(); err != nil {
		return err
	} else if user := app.UserFromContext(ctx); user == nil {
		return apperr.Errorf(apperr.EUNAUTHORIZED, "user is not authenticated")
	} else if contract.UserID != app.UserIDFromContext(ctx) {
		return apperr.Errorf(apperr.EUNAUTHORIZED, "contract is not owned by the authenticated user")
	} else if !user.CanCreateContract() {
		return apperr.Errorf(apperr.EFORBIDDEN, "user is not allowed to create a contract")
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO contracts (
			name,
			description,
			user_id,
			visibility,
			max_fuel,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`, contract.Name, contract.Description, contract.UserID, contract.Visibility, contract.MaxFuel, contract.CreatedAt, contract.UpdatedAt).Scan(&contract.ID); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to insert contract: %v", err)
	}

	return nil
}

// deleteContract deletes the contract with the given id.
// Return EFORBIDDEN if the user is not allowed to delete the contract.
// Return EUNAUTHORIZED if the contract is not owned by the authenticated user.
func deleteContract(ctx context.Context, tx *Tx, id int64) error {

	if contract, err := findContractByID(ctx, tx, id); err != nil {
		return err
	} else if contract.UserID != app.UserIDFromContext(ctx) {
		return apperr.Errorf(apperr.EUNAUTHORIZED, "contract is not owned by the authenticated user")
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM contracts WHERE id = $1`, id); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to delete contract: %v", err)
	}

	return nil
}

// findContractByID returns the contract with the given id.
func findContractByID(ctx context.Context, tx *Tx, id int64) (*entity.Contract, error) {

	c, _, err := findContracts(ctx, tx, service.ContractFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(c) == 0 {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "contract not found")
	}

	return c[0], nil
}

// findContracts returns a list of contracts filtered by the given options.
// Also returns the total count of contracts.
func findContracts(ctx context.Context, tx *Tx, filter service.ContractFilter) (_ entity.Contracts, n int, err error) {

	where, args := []string{"1 = 1"}, []interface{}{}

	counterParameter := 1

	if v := filter.ID; v != nil {
		where = append(where, fmt.Sprintf("id = $%d", counterParameter))
		args = append(args, *v)
		counterParameter++
	}
	if v := filter.Name; v != nil {
		where = append(where, fmt.Sprintf("name = $%d", counterParameter))
		args = append(args, *v)
		counterParameter++
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			name,
			description,
			user_id,
			visibility,
			max_fuel,
			created_at,
			updated_at,
			COUNT(*) OVER() as count
		FROM contracts
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset), args...)

	if err != nil {
		return nil, 0, apperr.Errorf(apperr.EINTERNAL, "failed to query contracts: %v", err)
	}
	defer rows.Close()

	contracts := make(entity.Contracts, 0)

	for rows.Next() {

		var contract entity.Contract

		if err := rows.Scan(
			&contract.ID,
			&contract.Name,
			&contract.Description,
			&contract.UserID,
			&contract.Visibility,
			&contract.MaxFuel,
			&contract.CreatedAt,
			&contract.UpdatedAt,
			&n,
		); err != nil {
			return nil, 0, apperr.Errorf(apperr.EINTERNAL, "failed to scan contract: %v", err)
		}

		contracts = append(contracts, &contract)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.Errorf(apperr.EINTERNAL, "failed to iterate over contracts: %v", err)
	}

	return contracts, n, nil
}

// makeRevision creates a new revision for the contract passed in.
func makeRevision(ctx context.Context, tx *Tx, revision *entity.Revision) error {

	revision.CreatedAt = tx.now

	if err := revision.Validate(); err != nil {
		return err
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO revisions (
			rev,
			version,
			contract_id,
			notes,
			code,
			compiled_code,
			max_fuel,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
	`, revision.Rev, revision.Version, revision.ContractID, revision.Notes, revision.Code, revision.CompiledCode, revision.MaxFuel, revision.CreatedAt).Scan(&revision.ID); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to insert revision: %v", err)
	}

	return nil
}

// updateContract updates the contract with the given id.
// Return EFORBIDDEN if the user is not allowed to update the contract.
// Return EUNAUTHORIZED if the contract is not owned by the authenticated user.
func updateContract(ctx context.Context, tx *Tx, id int64, upd service.ContractUpdate) (*entity.Contract, error) {

	contract, err := findContractByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if contract.UserID != app.UserIDFromContext(ctx) {
		return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "contract is not owned by the authenticated user")
	}

	if v := upd.Name; v != nil {
		contract.Name = *v
	}

	if v := upd.Description; v != nil {
		contract.Description = *v
	}

	contract.UpdatedAt = tx.now

	if err := contract.Validate(); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE contracts SET
			name = $1,
			description = $2,
			updated_at = $3
		WHERE id = $4
	`, contract.Name, contract.Description, contract.UpdatedAt, id); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to update contract: %v", err)
	}

	return contract, nil
}

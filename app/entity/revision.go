package entity

import (
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
)

// RevisionVersion indicates the version of revision management system.
type RevisionVersion string

// This is a list of all revisions versions.
const (
	AnchorageVersion RevisionVersion = "Anchorage"

	CurrentRevisionVersion RevisionVersion = AnchorageVersion
)

// RevisionNumber is the revision number of the entity.
type RevisionNumber uint

// Revisions is a slice of Revision.
type Revisions []*Revision

// Revision represents a revision of a contract.
// Each revision is a snapshot of the contract state.
type Revision struct {
	ID           int64           `json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	Rev          RevisionNumber  `json:"rev"`
	Version      RevisionVersion `json:"version"`
	ContractID   int64           `json:"contract_id"`
	Notes        string          `json:"notes"`
	CompiledCode []byte          `json:"-"`
	MaxFuel      Fuel            `json:"max_fuel"`

	Contract *Contract `json:"contract"`
}

// Validate validates the revision.
func (r *Revision) Validate() error {

	if r.Rev == 0 {

		return apperr.Errorf(apperr.EINVALID, "revision number cannot be zero")

	} else if r.ContractID == 0 {

		return apperr.Errorf(apperr.EINVALID, "contract id is required")

	} else if r.Version != CurrentRevisionVersion {

		return apperr.Errorf(apperr.EINVALID, "invalid revision version")

	} else if r.CreatedAt.IsZero() {

		return apperr.Errorf(apperr.EINVALID, "created at is required")

	} else if r.MaxFuel == 0 {

		return apperr.Errorf(apperr.EINVALID, "max fuel is required")

	} else if len(r.CompiledCode) == 0 {

		return apperr.Errorf(apperr.EINVALID, "compiled code is required")
	}

	return nil
}

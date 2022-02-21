package entity

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"gopkg.in/guregu/null.v4"
)

// Hash represents the hash an hash.
type Hash [32]byte

// RevisionVersion indicates the version of revision management system.
type RevisionVersion string

// This is a list of all revisions versions.
const (
	Anchorage RevisionVersion = "Anchorage"

	CurrentRevisionVersion RevisionVersion = Anchorage
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
	Rev          RevisionNumber  `json:"revision"`
	Version      RevisionVersion `json:"version"`
	ContractID   int64           `json:"contract_id"`
	Note         null.String     `json:"note"`
	Code         string          `json:"code"`
	CompiledCode []byte          `json:"-"`
	Hash         string          `json:"hash"`

	Contract *Contract `json:"contract"`
}

// CalculateHash calculates the hash of the revision.
func (r Revision) CalculateHash() (Hash, error) {

	if err := r.Validate(); err != nil {
		return Hash{}, err
	}

	var dataToHash string

	switch r.Version {
	case Anchorage:
		dataToHash = fmt.Sprintf("%d%s%d%s%d%s", r.ID, r.CreatedAt.Format(time.RFC3339), r.Rev, r.Version, r.ContractID, r.Code)
	default:
		return Hash{}, apperr.Errorf(apperr.EINVALID, "invalid revision version")
	}

	return sha256.Sum256([]byte(dataToHash)), nil
}

// Validate validates the revision.
func (r *Revision) Validate() error {

	if r.Rev == 0 {
		return apperr.Errorf(apperr.EINVALID, "revision number cannot be zero")
	}
	if r.Code == "" {
		return apperr.Errorf(apperr.EINVALID, "code is required")
	}
	if r.ContractID == 0 {
		return apperr.Errorf(apperr.EINVALID, "contract id is required")
	}
	if r.Version != CurrentRevisionVersion {
		return apperr.Errorf(apperr.EINVALID, "invalid revision version")
	}
	if r.CreatedAt.IsZero() {
		return apperr.Errorf(apperr.EINVALID, "created at is required")
	}

	return nil
}

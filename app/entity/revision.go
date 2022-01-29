package entity

import (
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"gopkg.in/guregu/null.v4"
)

// RevisionNumber is the revision number of the entity.
type RevisionNumber uint

// Revisions is a slice of Revision.
type Revisions []*Revision

// Revision represents a revision of a contract.
// Each revision is a snapshot of the contract state.
type Revision struct {
	ID        int64          `json:"id"`
	Rev       RevisionNumber `json:"revision"`
	CreatedAt time.Time      `json:"created_at"`
	Note      null.String    `json:"note"`
	Code      string         `json:"code"`

	Contract *Contract `json:"contract"`
}

// Validate validates the revision.
func (r *Revision) Validate() error {

	// You think i forgot to validate the revision number?
	// I don't.
	// Revision number is generated by the database.
	// :D

	if r.Code == "" {
		return apperr.Errorf(apperr.EINVALID, "code is required")
	}

	return nil
}
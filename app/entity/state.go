package entity

import "github.com/music-gang/music-gang-api/app/apperr"

// State represents the state value.
type State []byte

// ContractState represents the state of the contract at a given revision for specific user.
// ContractState enables contract to persist its state during different executions.
// The state should not be shared between users and different revisions.
type ContractState struct {
	ID         int64 `json:"id"`
	RevisionID int64 `json:"revision_id"`
	State      State `json:"state"`
	UserID     int64 `json:"user_id"`
	CreatedAt  int64 `json:"created_at"`
	UpdatedAt  int64 `json:"updated_at"`

	User     *User     `json:"user"`
	Revision *Revision `json:"revision"`
}

// Validate validates the state.
func (s ContractState) Validate() error {

	if s.RevisionID == 0 {
		return apperr.Errorf(apperr.EINVALID, "revision id is required")
	}

	if s.UserID == 0 {
		return apperr.Errorf(apperr.EINVALID, "user id is required")
	}

	return nil
}

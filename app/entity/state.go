package entity

import (
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/util"
)

// StateValue represents the state value.
type StateValue map[string]any

// NewStateFromBytes creates a state from bytes.
func NewStateFromBytes(b []byte) (StateValue, error) {
	s := make(StateValue)
	if b != nil {
		if err := util.FromBytes(b, s); err != nil {
			return nil, apperr.Errorf(apperr.EINTERNAL, "Error while converting bytes to state: %s", err.Error())
		}
	}
	return s, nil
}

// Bytes converts the state to bytes.
func (s StateValue) Bytes() ([]byte, error) {
	b, err := util.ToBytes(s)
	if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "Error while converting state to bytes: %s", err.Error())
	}
	return b, nil
}

// State represents the state of the contract at a given revision for specific user.
// State enables contract to persist its state during different executions.
// The state should not be shared between users and different revisions.
type State struct {
	ID         int64      `json:"id"`
	RevisionID int64      `json:"revision_id"`
	Value      StateValue `json:"value"`
	UserID     int64      `json:"user_id"`
	CreatedAt  int64      `json:"created_at"`
	UpdatedAt  int64      `json:"updated_at"`

	User     *User     `json:"user"`
	Revision *Revision `json:"revision"`
}

// Validate validates the state.
func (s State) Validate() error {

	if s.RevisionID == 0 {
		return apperr.Errorf(apperr.EINVALID, "revision id is required")
	}

	if s.UserID == 0 {
		return apperr.Errorf(apperr.EINVALID, "user id is required")
	}

	return nil
}

package entity

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
)

const (
	// EmptyState is the empty state.
	EmptyState = `{}`
)

var _ driver.Valuer = StateValue{}
var _ sql.Scanner = (*StateValue)(nil)

// StateValue represents the state value.
type StateValue map[string]any

// Value implements driver.Valuer
func (s StateValue) Value() (driver.Value, error) {
	v, err := json.Marshal(s)
	if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "Error while converting state to Value: %s", err.Error())
	}
	return v, nil
}

// Scan implements sql.Scanner
func (s *StateValue) Scan(src any) error {
	b, ok := src.([]byte)
	if !ok {
		return apperr.Errorf(apperr.EINVALID, "invalid state value")
	}
	return json.Unmarshal(b, &s)
}

// NewStateFromBytes creates a state from bytes.
// If b is nil, it returns an empty state.
func NewStateFromBytes(b []byte) (StateValue, error) {
	s := make(StateValue)
	if b == nil {
		b = []byte(EmptyState)
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "Error while converting bytes to state: %s", err.Error())
	}
	return s, nil
}

// State represents the state of the contract at a given revision for specific user.
// State enables contract to persist its state during different executions.
// The state should not be shared between users and different revisions.
type State struct {
	ID         int64      `json:"id"`
	RevisionID int64      `json:"revision_id"`
	Value      StateValue `json:"value"`
	UserID     int64      `json:"user_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

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

	if s.Value == nil {
		return apperr.Errorf(apperr.EINVALID, "value cannot be nil")
	}

	return nil
}

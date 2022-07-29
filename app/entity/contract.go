package entity

import (
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
)

// Visibility consts for the visibility of the contract.
const (
	VisibilityPrivate = "private"
	VisibilityPublic  = "public"
)

// Visibility defines the visibility of a contract.
type Visibility string

// Validate validates the visibility.
func (v Visibility) Validate() error {
	switch v {
	case
		VisibilityPrivate,
		VisibilityPublic:
		return nil
	default:
		return apperr.Errorf(apperr.EINVALID, "invalid visibility")
	}
}

// Contracts represents a list of contracts.
type Contracts []*Contract

// Contract represents a contract.
// The contract is a cloud function that is executed on a server, deployed by users;
// The contract can have multiple revisions.
type Contract struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	UserID      int64      `json:"user_id"`
	Visibility  Visibility `json:"visibility"`
	MaxFuel     Fuel       `json:"max_fuel"` // The maximum amount of fuel that can be burned from the contract.
	Stateful    bool       `json:"stateful"` // Enables the contract to persist its state during different executions (of same revision).
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// avoid to access this field directly because it can be nil, use the Revision method UnwrapRevision instead.
	LastRevision *Revision `json:"last_revision"`
	User         *User     `json:"user"`
}

// MaxExecutionTime returns the maximum execution time of the contract.
// MaxExecutionTime is based on max fuel compared with fuelAmountTable.
func (c *Contract) MaxExecutionTime() time.Duration {
	return MaxExecutionTimeFromFuel(c.MaxFuel)
}

// Validate validates the contract.
func (c *Contract) Validate() error {

	if c.Name == "" {
		return apperr.Errorf(apperr.EINVALID, "contract name is required")
	}

	if c.UserID == 0 {
		return apperr.Errorf(apperr.EINVALID, "User ID is required")
	}

	if c.MaxFuel == 0 {
		return apperr.Errorf(apperr.EINVALID, "Max fuel is required")
	}

	if err := c.Visibility.Validate(); err != nil {
		return err
	}

	return nil
}

// UnwrapRevision returns the last revision of the contract if it exists, otherwise error is returned.
func (c *Contract) UnwrapRevision() (*Revision, error) {
	if c.LastRevision == nil {
		return nil, apperr.Errorf(apperr.ENOTFOUND, "no revision found")
	}
	return c.LastRevision, nil
}

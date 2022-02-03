package entity

import (
	"github.com/music-gang/music-gang-api/app/apperr"
	"gopkg.in/guregu/null.v4"
)

// Visibility consts for the visibility of the contract.
const (
	VisibilityPrivate = iota
	VisibilityPublic
)

// Visibility defines the visibility of a contract.
type Visibility int

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
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Description null.String `json:"description"`
	UserID      null.Int    `json:"user_id"`
	Visibility  Visibility  `json:"visibility"`
	MaxFuel     Fuel        `json:"max_fuel"` // The maximum amount of fuel that can be burned from the contract.

	LastRevision *Revision `json:"last_revision"`
	User         *User     `json:"user"`
}

// Validate validates the contract.
func (c *Contract) Validate() error {

	if c.Name == "" {
		return apperr.Errorf(apperr.EINVALID, "contract name is required")
	}

	if c.UserID.Valid && c.UserID.Int64 == 0 {
		return apperr.Errorf(apperr.EINVALID, "User ID cannot be empty if provided")
	}

	if err := c.Visibility.Validate(); err != nil {
		return err
	}

	return nil
}

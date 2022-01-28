package entity

import "time"

type Contract struct {
	ID             int64
	Name           string
	LastFuelUsed   Fuel
	LastDuration   time.Duration
	LastRevisionID int64
	UserID         int64

	Revision *Revision
	User     *User
	Args     []interface{}
}

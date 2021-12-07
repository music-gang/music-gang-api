package entity

import (
	"time"

	"github.com/music-gang/music-gang-api/app"
	"gopkg.in/guregu/null.v4"
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt null.Time `json:"-"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"-"`
}

type Users []*User

func (u *User) Validate() error {

	if u.Name == "" {
		return app.Errorf(app.EINVALID, "name is required")
	}

	if u.Username == "" {
		return app.Errorf(app.EINVALID, "username is required")
	}

	if u.Email == "" {
		return app.Errorf(app.EINVALID, "email is required")
	}

	return nil
}

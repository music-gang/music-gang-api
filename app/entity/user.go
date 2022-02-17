package entity

import (
	"strings"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"gopkg.in/guregu/null.v4"
)

const (
	// UserNameInvalidCharacters defines all invalid characters for a user name.
	UserNameInvalidCharacters = "!@#$%^&*()+=[]{}|\\;:'\"<>,/?`~"
)

// User represents a user in the system. Users are typically created via OAuth
// using the AuthService but users can also be created directly for testing.
type User struct {
	ID        int64       `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Email     null.String `json:"email"`
	Name      string      `json:"name"`
	Password  null.String `json:"-"`

	// List of associated OAuth authentication objects.
	// Currently only GitHub is supported so there should only be a maximum of one.
	Auths []*Auth `json:"auths"`
}

// Users is a list of users.
type Users []*User

// CanCreateContract returns if user is authorized to created contracts.
func (u *User) CanCreateContract() bool {
	return u.ID != 0
}

// Validate returns an error if the user contains invalid fields.
// This only performs basic validation.
func (u *User) Validate() error {

	if u.Name == "" {
		return apperr.Errorf(apperr.EINVALID, "name is required")
	}

	// check if name contains whitespaces. This is not allowed.
	if strings.Contains(u.Name, " ") {
		return apperr.Errorf(apperr.EINVALID, "name cannot contain whitespaces")
	}

	// check if name contains invalid characters. This is not allowed.
	if strings.ContainsAny(u.Name, UserNameInvalidCharacters) {
		return apperr.Errorf(apperr.EINVALID, "name cannot contain invalid characters")
	}

	if u.Email.Valid && u.Email.String == "" {
		return apperr.Errorf(apperr.EINVALID, "email cannot be empty if provided")
	}

	return nil
}

// AvatarURL returns a URL to the avatar image for the user.
// This loops over all auth providers to find the first available avatar.
// Currently only GitHub is supported. Returns blank string if no avatar URL available.
func (u *User) AvatarURL(size int) string {
	for _, auth := range u.Auths {
		if s := auth.AvatarURL(size); s != "" {
			return s
		}
	}
	return ""
}

package models

import (
	"errors"

	"code.google.com/p/go.crypto/bcrypt"
)

// Constants representing the various roles a user may possess
const (
	RoleGuest = iota
	RoleUser
	RoleAdmin
)

var (
	// ErrInvalidPassword is returned when password authentication fails for a
	// specified User.
	ErrInvalidPassword = errors.New("invalid password")
)

// Ensure User implements Validator.
var _ Validator = &User{}

// User represents a user of the application.
type User struct {
	ID          uint64 `db:"id" json:"id"`
	Username    string `db:"username" json:"username"`
	Password    string `db:"password" json:"password,omitempty"`
	RoleID      uint8  `db:"role_id" json:"roleId"`
	LastFMToken string `db:"lastfm_token" json:"-"`
}

// SetPassword hashes the input password using bcrypt, storing the password
// within the receiving User struct.
func (u *User) SetPassword(password string) error {
	// Check for empty password
	if password == "" {
		return &EmptyFieldError{
			Field: "password",
		}
	}

	// Generate password hash using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)

	return nil
}

// TryPassword attempts to verify the input password against the receiving User's
// current password.
func (u *User) TryPassword(password string) error {
	// Attempt to hash password
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	// Check for bcrypt-specific password failure, return more generic failure
	// (other packages should not have to import or know about bcrypt to know
	// the password was incorrect)
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrInvalidPassword
	}

	// Return other errors, or no error
	return err
}

// SQLReadFields returns the correct field order to scan SQL row results into the
// receiving User struct.
func (u *User) SQLReadFields() []interface{} {
	return []interface{}{
		&u.ID,
		&u.Username,
		&u.Password,
		&u.RoleID,
		&u.LastFMToken,
	}
}

// SQLWriteFields returns the correct field order for SQL write actions (such as
// insert or update), for the receiving User struct.
func (u *User) SQLWriteFields() []interface{} {
	return []interface{}{
		u.Username,
		u.Password,
		u.RoleID,
		u.LastFMToken,

		// Last argument for WHERE clause
		u.ID,
	}
}

// Validate verifies that all fields for the receiving User struct contain
// valid input.
func (u *User) Validate() error {
	// Check for required fields
	if u.Username == "" {
		return &EmptyFieldError{
			Field: "username",
		}
	}
	if u.Password == "" {
		return &EmptyFieldError{
			Field: "password",
		}
	}

	return nil
}

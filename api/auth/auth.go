package auth

import (
	"net/http"
)

// AuthMethod represents a method of authenticating with the API
type AuthMethod interface {
	Authenticate(*http.Request) bool
}

// BcryptAuth represents the bcrypt authentication method, used to log in to the API
type BcryptAuth struct{}

// Authenticate uses the bcrypt authentication method to log in to the API
func (a BcryptAuth) Authenticate(req *http.Request) bool {
	// TODO: implement this method
	return true
}

// APIAuth represents the standard API authentication method, used for all other API calls
type APIAuth struct{}

// Authenticate uses the standard API authentication method for any calls outside of login
func (a APIAuth) Authenticate(req *http.Request) bool {
	// TODO: implement this method
	return true
}

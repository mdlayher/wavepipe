package auth

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/mdlayher/wavepipe/data"
)

// SimpleAuth represents the simple authentication method, which can be used by local clients
type SimpleAuth struct{}

// Authenticate uses the simple authentication method to log in to the API, returning
// a session user and a pair of client/server errors
func (a SimpleAuth) Authenticate(req *http.Request) (*data.User, *data.Session, error, error) {
	// Username for authentication
	var username string

	// Check for empty authorization header
	if req.Header.Get("Authorization") == "" {
		// If no header, check for credentials via querystring parameters
		query := req.URL.Query()
		username = query.Get("u")
	} else {
		// Fetch credentials from HTTP Basic auth
		tempUsername, _, err := basicCredentials(req.Header.Get("Authorization"))
		if err != nil {
			return nil, nil, err, nil
		}

		// Copy credentials
		username = tempUsername
	}

	// Check if username is blank
	if username == "" {
		return nil, nil, ErrNoUsername, nil
	}

	// Attempt to load user by username
	user := new(data.User)
	user.Username = username
	if err := user.Load(); err != nil {
		// Check for invalid user
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("invalid username"), nil
		}

		// Server error
		return nil, nil, nil, err
	}

	// No errors, return session user, but no session because one does not exist yet
	return user, nil, nil, nil
}

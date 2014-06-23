package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"os"

	"github.com/mdlayher/wavepipe/data"
)

// simpleAuthenticate uses the simple authentication method to log in to the API, returning
// a session user and a pair of client/server errors.
func simpleAuthenticate(req *http.Request) (*data.User, *data.Session, error, error) {
	// Verify that SimpleAuth was triggered in debug mode
	if os.Getenv("WAVEPIPE_DEBUG") != "1" {
		return nil, nil, nil, errors.New("not in debug mode")
	}

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

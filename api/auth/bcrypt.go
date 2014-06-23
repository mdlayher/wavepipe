package auth

import (
	"database/sql"
	"net/http"

	"github.com/mdlayher/wavepipe/data"

	"code.google.com/p/go.crypto/bcrypt"
)

// bcryptAuthenticate uses the bcrypt authentication method to log in to the API, returning
// a session user and a pair of client/server errors
func bcryptAuthenticate(req *http.Request) (*data.User, *data.Session, error, error) {
	// Username and password for authentication
	var username string
	var password string

	// Check for empty authorization header
	if req.Header.Get("Authorization") == "" {
		// If no header, check for credentials via POST parameters
		username = req.PostFormValue("username")
		password = req.PostFormValue("password")
	} else {
		// Fetch credentials from HTTP Basic auth
		tempUsername, tempPassword, err := basicCredentials(req.Header.Get("Authorization"))
		if err != nil {
			return nil, nil, err, nil
		}

		// Copy credentials
		username = tempUsername
		password = tempPassword
	}

	// Check if either credential is blank
	if username == "" {
		return nil, nil, ErrNoUsername, nil
	} else if password == "" {
		return nil, nil, ErrNoPassword, nil
	}

	// Attempt to load user by username
	user := new(data.User)
	user.Username = username
	if err := user.Load(); err != nil {
		// Check for invalid user
		if err == sql.ErrNoRows {
			return nil, nil, ErrInvalidUsername, nil
		}

		// Server error
		return nil, nil, nil, err
	}

	// Compare input password with bcrypt password, checking for errors
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		// Mismatch password
		return nil, nil, ErrInvalidPassword, nil
	} else if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
		// Return server error
		return nil, nil, nil, err
	}

	// No errors, return session user, but no session because one does not exist yet
	return user, nil, nil, nil
}

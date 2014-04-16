package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"code.google.com/p/go.crypto/bcrypt"
)

var (
	// ErrNoUsername is returned when no username is provided on login
	ErrNoUsername = errors.New("no username provided")
	// ErrNoPassword is returned when no password is provided on login
	ErrNoPassword = errors.New("no password provided")
)

// AuthMethod represents a method of authenticating with the API
type AuthMethod interface {
	Authenticate(*http.Request) (error, error)
}

// BcryptAuth represents the bcrypt authentication method, used to log in to the API
type BcryptAuth struct{}

// Authenticate uses the bcrypt authentication method to log in to the API, returning
// a pair of client/server errors
func (a BcryptAuth) Authenticate(req *http.Request) (error, error) {
	// Username and password for authentication
	var username string
	var password string

	// Check for empty authorization header
	if req.Header.Get("Authorization") == "" {
		// If no header, check for credentials via querystring parameters
		query := req.URL.Query()
		username = query.Get("u")
		password = query.Get("p")
	} else {
		// Fetch credentials from HTTP Basic auth
		tempUsername, tempPassword, err := basicCredentials(req.Header.Get("Authorization"))
		if err != nil {
			return err, nil
		}

		// Copy credentials
		username = tempUsername
		password = tempPassword
	}

	// Check if either credential is blank
	if username == "" {
		return ErrNoUsername, nil
	} else if password == "" {
		return ErrNoPassword, nil
	}

	// Generate a fake bcrypt hash
	hash, err := bcrypt.GenerateFromPassword([]byte("test"), 10)
	if err != nil {
		return nil, err
	}

	// Load user by username
	// TODO: create User model and database calls
	user := struct{
		Username string
		Password string
	}{
		"test",
		string(hash),
	}

	// Compare input password with bcrypt password, checking for errors
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		// Mismatch password
		return errors.New("invalid password"), nil
	} else if err != bcrypt.ErrMismatchedHashAndPassword {
		// Return server error
		return nil, err
	}

	// No errors
	return nil, nil
}

// APIAuth represents the standard API authentication method, used for all other API calls
type APIAuth struct{}

// Authenticate uses the standard API authentication method for any calls outside of login
func (a APIAuth) Authenticate(req *http.Request) (error, error) {
	// TODO: implement this method
	return nil, nil
}

// basicCredentials returns HTTP Basic authentication credentials from a header
func basicCredentials(header string) (string, string, error) {
	// No headed provided
	if header == "" {
		return "", "", errors.New("empty HTTP Basic header")
	}

	// Ensure valid format
	basic := strings.Split(header, " ")
	if basic[0] != "Basic" {
		return "", "", errors.New("invalid HTTP Basic header")
	}

	// Decode base64'd username:password pair
	buf, err := base64.URLEncoding.DecodeString(basic[1])
	if err != nil {
		return "", "", errors.New("invalid HTTP Basic header")
	}

	// Split into username/password
	pair := strings.Split(string(buf), ":")
	return pair[0], pair[1], nil
}

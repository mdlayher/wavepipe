package auth

import (
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/mdlayher/wavepipe/data"
	"github.com/mdlayher/wavepipe/env"
)

var (
	// ErrNoUsername is returned when no username is provided on login
	ErrNoUsername = errors.New("no username provided")
	// ErrNoPassword is returned when no password is provided on login
	ErrNoPassword = errors.New("no password provided")
	// ErrInvalidUsername is returned when an invalid username is provided on login
	ErrInvalidUsername = errors.New("invalid username")
	// ErrInvalidPassword is returned when an invalid password is provided on login
	ErrInvalidPassword = errors.New("invalid password")

	// ErrNoToken is returned when no API token is provided on all other API calls
	ErrNoToken = errors.New("no token provided")
	// ErrInvalidToken is returned when an invalid API token is provided on all other API calls
	ErrInvalidToken = errors.New("invalid token")
	// ErrSessionExpired is returned when the session is expired
	ErrSessionExpired = errors.New("session expired")

	// ErrEmptyBasic is returned when a blank HTTP Basic authentication header is passed
	ErrEmptyBasic = errors.New("empty HTTP Basic header")
	// ErrInvalidBasic is returned when an invalid HTTP Basic authentication header is passed
	ErrInvalidBasic = errors.New("invalid HTTP Basic header")
)

// AuthenticatorFunc is an adapter which allows any function with the appropriate signature to act as an
// authentication function for the wavepipe API
type AuthenticatorFunc func(*http.Request) (*data.User, *data.Session, error, error)

// Authenticate invokes an AuthenticatorFunc and returns a user, its session, a client error, and a server error
func (f AuthenticatorFunc) Authenticate(req *http.Request) (*data.User, *data.Session, error, error) {
	return f(req)
}

// Factory generates the appropriate authorization method by using input parameters
func Factory(path string) AuthenticatorFunc {
	// Check for debug mode, and if it's set, automatically use the Simple method
	if env.IsDebug() {
		log.Println("api: warning: authenticating user in debug mode")
		return simpleAuthenticate
	}

	// Check for request to emulated Subsonic API, which is authenticated using
	// its own, special method which outputs XML
	if strings.HasPrefix(path, "/subsonic") {
		return subsonicAuthenticate
	}

	// Check if path does not reside under the /api, meaning it is unauthenticated
	if !strings.HasPrefix(path, "/api") {
		return nilAuthenticate
	}

	// Strip any trailing slashes from the path
	path = strings.TrimRight(path, "/")

	// Check for request to API root (/api, /api/vX), which is unauthenticated
	if path == "/api" || (strings.HasPrefix(path, "/api/v") && len(path) == 7) {
		return nilAuthenticate
	}

	// Check for a login request: /api/vX/login, use bcrypt authenticator
	// Note: length check added to prevent this from happening on /api/VX/lastfm/login
	if len(path) == 13 && strings.HasPrefix(path, "/api/v") && strings.HasSuffix(path, "/login") {
		return bcryptAuthenticate
	}

	// All other situations, use the token authenticator
	return tokenAuthenticate
}

// basicCredentials returns HTTP Basic authentication credentials from a header
func basicCredentials(header string) (string, string, error) {
	// No headed provided
	if header == "" {
		return "", "", ErrEmptyBasic
	}

	// Ensure 2 elements
	basic := strings.Split(header, " ")
	if len(basic) != 2 {
		return "", "", ErrInvalidBasic
	}

	// Ensure valid format
	if basic[0] != "Basic" {
		return "", "", ErrInvalidBasic
	}

	// Decode base64'd username:password pair
	buf, err := base64.URLEncoding.DecodeString(basic[1])
	if err != nil {
		return "", "", ErrInvalidBasic
	}

	// Split into username/password
	pair := strings.SplitN(string(buf), ":", 2)
	if len(pair) < 2 {
		return "", "", ErrInvalidBasic
	}

	return pair[0], pair[1], nil
}

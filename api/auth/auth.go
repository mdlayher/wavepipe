package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/mdlayher/wavepipe/data"
)

var (
	// ErrNoUsername is returned when no username is provided on login
	ErrNoUsername = errors.New("no username provided")
	// ErrNoPassword is returned when no password is provided on login
	ErrNoPassword = errors.New("no password provided")
)

// AuthMethod represents a method of authenticating with the API
type AuthMethod interface {
	Authenticate(*http.Request) (*data.User, *data.Session, error, error)
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

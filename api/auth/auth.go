package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mdlayher/wavepipe/data"

	"github.com/willf/bloom"
)

var (
	// ErrNoUsername is returned when no username is provided on login
	ErrNoUsername = errors.New("no username provided")
	// ErrNoPassword is returned when no password is provided on login
	ErrNoPassword = errors.New("no password provided")

	// ErrInvalidPublicKey is returned when an invalid public keyis used to access the API
	ErrInvalidPublicKey = errors.New("no such public key")
	// ErrNoSignature is returned when no API signature is provided on all other API calls
	ErrNoSignature = errors.New("no signature provided")
	// ErrNoToken is returned when no API token is provided on all other API calls
	ErrNoToken = errors.New("no token provided")
	// ErrInvalidSignature is returned when a mismatched API signature is provided
	ErrInvalidSignature = errors.New("invalid signature provided")
	// ErrMalformedSignature is returned when a malformed API signature is provided
	ErrMalformedSignature = errors.New("malformed signature provided")
	// ErrRepeatedRequest is returned when an API nonce is re-used
	ErrRepeatedRequest = errors.New("repeated nonce provided")
	// ErrSessionExpired is returned when the session is expired
	ErrSessionExpired = errors.New("session expired")
)

// NonceFilter is a bloom filter containing all nonce values seen in the past 24 hours.
// The filter is cleared every 24 hours, because this should be a reasonable enough time to
// prevent replay attacks.
var NonceFilter = bloom.New(20000, 5)

// AuthMethod represents a method of authenticating with the API
type AuthMethod interface {
	Authenticate(*http.Request) (*data.User, *data.Session, error, error)
}

// Factory generates the appropriate authorization method by using input parameters
func Factory(path string) AuthMethod {
	// Check if path does not reside under the /api, meaning it is unauthenticated
	if !strings.HasPrefix(path, "/api") {
		return nil
	}

	// Strip any trailing slashes from the path
	path = strings.TrimRight(path, "/")

	// Check for request to API root (/api, /api/vX), which is unauthenticated
	if path == "/api" || (strings.HasPrefix(path, "/api/v") && len(path) == 7) {
		return nil
	}

	// Check for request to emulated Subsonic API, which is authenticated elsewhere
	// due to many differences from the wavepipe API
	if strings.HasPrefix(path, "/subsonic") {
		return nil
	}

	// Check for a login request: /api/vX/login, use bcrypt authenticator
	if strings.HasPrefix(path, "/api/v") && strings.HasSuffix(path, "/login") {
		return new(BcryptAuth)
	}

	// Check for debug mode, and if it's set, automatically use the Simple method
	if os.Getenv("WAVEPIPE_DEBUG") == "1" {
		log.Println("api: warning: authenticating user in debug mode")
		return new(SimpleAuth)
	}

	// All other situations, use the token authenticator
	return new(TokenAuth)
}

// apiSignature generates a HMAC-SHA1 signature for use with the API
func apiSignature(public string, nonce string, method string, resource string, secret string) (string, error) {
	// Generate API signature string
	signString := fmt.Sprintf("%s-%s-%s-%s", public, nonce, method, resource)

	// Calculate HMAC-SHA1 signature from string, using API secret
	mac := hmac.New(sha1.New, []byte(secret))
	if _, err := mac.Write([]byte(signString)); err != nil {
		return "", err
	}

	// Return hex signature
	return fmt.Sprintf("%x", mac.Sum(nil)), nil
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
	pair := strings.SplitN(string(buf), ":", 2)
	if len(pair) < 2 {
		return "", "", errors.New("invalid HTTP Basic username/password combination")
	}

	return pair[0], pair[1], nil
}

package auth

import (
	"crypto/hmac"
	"net/http"
	"strings"
	"time"

	"github.com/mdlayher/wavepipe/data"
)

// HMACAuth represents the standard API authentication method, HMAC-SHA1
type HMACAuth struct{}

// Authenticate uses the HMAC-SHA1 authentication scheme for any calls outside of login
func (a HMACAuth) Authenticate(req *http.Request) (*data.User, *data.Session, error, error) {
	// Public key, nonce, and API signature for authentication
	var publicKey string
	var nonce string
	var signature string

	// Check for empty authorization header
	if req.Header.Get("Authorization") == "" {
		// If no header, check for credentials via querystring parameter
		packedSig := req.URL.Query().Get("s")

		// Check for empty signature
		if packedSig == "" {
			return nil, nil, ErrNoSignature, nil
		}

		// Attempt to split the packed signature into components
		triple := strings.Split(packedSig, ":")
		if len(triple) < 3 {
			return nil, nil, ErrMalformedSignature, nil
		}

		// Copy components
		publicKey = triple[0]
		nonce = triple[1]
		signature = triple[2]
	} else {
		// Fetch credentials from HTTP Basic auth
		tempPublicKey, tempNonceSignature, err := basicCredentials(req.Header.Get("Authorization"))
		if err != nil {
			return nil, nil, err, nil
		}

		// Split second component
		pair := strings.Split(tempNonceSignature, ":")
		if len(pair) < 2 {
			return nil, nil, ErrMalformedSignature, nil
		}

		// Copy components
		publicKey = tempPublicKey
		nonce = pair[0]
		signature = pair[1]
	}

	// Check if nonce previously used, add it if it is not, to prevent replay attacks
	// note: bloom filter may report false positives, but better safe than sorry
	if NonceFilter.TestAndAdd([]byte(nonce)) {
		return nil, nil, ErrRepeatedRequest, nil
	}

	// Attempt to load session by its public key
	session := new(data.Session)
	session.PublicKey = publicKey
	if err := session.Load(); err != nil {
		return nil, nil, nil, err
	}

	// Check if session is expired, delete it if it is
	if session.Expire <= time.Now().Unix() {
		if err := session.Delete(); err != nil {
			return nil, nil, nil, err
		}

		// Report failure to user
		return nil, nil, ErrSessionExpired, nil
	}

	// Generate API signature
	expected, err := apiSignature(session.PublicKey, nonce, req.Method, req.URL.Path, session.SecretKey)
	if err != nil {
		return nil, nil, nil, err
	}

	// Verify that HMAC signature is correct
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return nil, nil, ErrInvalidSignature, nil
	}

	// Update API session expiration time
	session.Expire = time.Now().Add(24 * time.Hour).Unix()
	if err := session.Update(); err != nil {
		return nil, nil, nil, err
	}

	// Load user for session
	user := new(data.User)
	user.ID = session.UserID
	if err := user.Load(); err != nil {
		return nil, nil, nil, err
	}

	// Return user and session
	return user, session, nil, nil
}

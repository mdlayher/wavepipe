package auth

import (
	"errors"
	"net/http"

	"github.com/mdlayher/wavepipe/data"
)

// HMACAuth represents the standard API authentication method, HMAC-SHA1
type HMACAuth struct{}

// Authenticate uses the HMAC-SHA1 authentication scheme for any calls outside of login
func (a HMACAuth) Authenticate(req *http.Request) (*data.User, *data.Session, error, error) {
	// TODO: implement this method
	return nil, nil, nil, errors.New("method not implemented")
}

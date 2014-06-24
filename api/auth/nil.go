package auth

import (
	"net/http"

	"github.com/mdlayher/wavepipe/data"
)

// nilAuthenticate uses the nil authentication method to log in to the API, returning
// a session user and a pair of client/server errors.  This method is used for routes
// which are unauthenticated.
func nilAuthenticate(req *http.Request) (*data.User, *data.Session, error, error) {
	// Return all parameters as blank
	return new(data.User), new(data.Session), nil, nil
}

package auth

import (
	"net/http"

	"github.com/mdlayher/wavepipe/data"
	"github.com/mdlayher/wavepipe/subsonic"
)

// SubsonicAuth represents the Subsonic authentication method, which is used ONLY for the Subsonic
// emulation layer
type SubsonicAuth struct{}

// Authenticate uses the Subsonic authentication method to log in to the API, returning
// only a pair of client/server errors
func (a SubsonicAuth) Authenticate(req *http.Request) (*data.User, *data.Session, error, error) {
	// Check for required credentials via querystring
	query := req.URL.Query()
	username := query.Get("u")
	password := query.Get("p")

	// Check if username or password is blank
	if username == "" || password == "" {
		return nil, nil, subsonic.ErrBadCredentials, nil
	}

	// Check for Subsonic version
	version := query.Get("v")
	if version == "" {
		return nil, nil, subsonic.ErrMissingParameter, nil
	}

	// TODO: add authentication logic here

	// No errors, return no user or session
	return nil, nil, nil, nil
}

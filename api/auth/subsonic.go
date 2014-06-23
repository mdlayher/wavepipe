package auth

import (
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/mdlayher/wavepipe/data"
	"github.com/mdlayher/wavepipe/subsonic"
)

// subsonicAuthenticate uses the Subsonic authentication method to log in to the API, returning
// only a pair of client/server errors
func subsonicAuthenticate(req *http.Request) (*data.User, *data.Session, error, error) {
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

	// TODO: reevaluate this strategy in the future, but for now, we will use a user's wavepipe session
	// TODO: key as their Subsonic password.  This will mean that the username and session key are passed on
	// TODO: every request, but also means that no more database schema must be added for Subsonic authentication.

	// Check for "enc:" prefix, specifying a hex-encoded password
	if strings.HasPrefix(password, "enc:") {
		// Decode hex string
		out, err := hex.DecodeString(password[4:])
		if err != nil {
			return nil, nil, nil, err
		}

		password = string(out)
	}

	// Attempt to load session by key passed via Subsonic password parameter
	session := new(data.Session)
	session.Key = password
	if err := session.Load(); err != nil {
		// Check for invalid user
		if err == sql.ErrNoRows {
			return nil, nil, subsonic.ErrBadCredentials, nil
		}

		// Server error
		return nil, nil, nil, err
	}

	// Attempt to load associated user by username from Subsonic username parameter
	user := new(data.User)
	user.Username = username
	if err := user.Load(); err != nil {
		// Server error
		return nil, nil, nil, err
	}

	// Update session expiration date by 1 week
	session.Expire = time.Now().Add(7 * 24 * time.Hour).Unix()
	if err := session.Update(); err != nil {
		return nil, nil, nil, err
	}

	// No errors, return no user or session because the emulated Subsonic API is read-only
	return nil, nil, nil, nil
}

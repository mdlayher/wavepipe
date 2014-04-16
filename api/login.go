package api

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// LoginResponse represents the JSON response for /api/logins
type LoginResponse struct {
	Error   *Error        `json:"error"`
	Session *data.Session `json:"session"`
}

// GetLogin creates a new session on the wavepipe API, and returns a HTTP status and JSON
func GetLogin(r render.Render, req *http.Request, sessionUser *data.User, params martini.Params) {
	// Output struct for logins request
	res := LoginResponse{}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.Error = new(Error)
			res.Error.Code = 400
			res.Error.Message = "unsupported API version: " + version
			r.JSON(400, res)
			return
		}
	}

	// Generate a new API session for this user, with optional specified session name
	// via "c" query parameter
	session, err := sessionUser.CreateSession(req.URL.Query().Get("c"))
	if err != nil {
		log.Println(err)

		res.Error = new(Error)
		res.Error.Code = 500
		res.Error.Message = "server error"
		r.JSON(500, res)
		return
	}

	// Build response
	res.Error = nil
	res.Session = session

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}

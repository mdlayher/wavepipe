package api

import (
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// LoginResponse represents the JSON response for /api/logins
type LoginResponse struct {
	Error   *Error `json:"error"`
	Session string `json:"session"`
}

// GetLogin creates a new session on the wavepipe API, and returns a HTTP status and JSON
func GetLogin(r render.Render, req *http.Request, params martini.Params) {
	// Output struct for logins request
	res := LoginResponse{}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.Error = new(Error)
			res.Error.Code = http.StatusBadRequest
			res.Error.Message = "unsupported API version: " + version
			r.JSON(http.StatusBadRequest, res)
			return
		}
	}

	// TODO: implement session generation logic

	// Build response
	res.Error = nil
	res.Session = "abcdef0123456789"

	// HTTP 200 OK with JSON
	r.JSON(http.StatusOK, res)
	return
}

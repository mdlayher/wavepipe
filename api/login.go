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
	render  render.Render `json:"-"`
}

// RenderError renders a JSON error message with the specified HTTP status code and message
func (l *LoginResponse) RenderError(code int, message string) {
	// Nullify all other fields
	l.Session = nil

	// Generate error
	l.Error = new(Error)
	l.Error.Code = code
	l.Error.Message = message

	// Render with specified HTTP status code
	l.render.JSON(code, l)
}

// ServerError is a shortcut to render a HTTP 500 with generic "server error" message
func (l *LoginResponse) ServerError() {
	l.RenderError(500, "server error")
	return
}

// GetLogin creates a new session on the wavepipe API, and returns a HTTP status and JSON
func GetLogin(r render.Render, req *http.Request, sessionUser *data.User, params martini.Params) {
	// Output struct for logins request
	res := LoginResponse{render: r}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Generate a new API session for this user, with optional specified session name
	// via "c" query parameter
	session, err := sessionUser.CreateSession(req.URL.Query().Get("c"))
	if err != nil {
		log.Println(err)
		res.ServerError()
		return
	}

	// Build response
	res.Error = nil
	res.Session = session

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}

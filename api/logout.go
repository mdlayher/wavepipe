package api

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// LogoutResponse represents the JSON response for /api/logouts
type LogoutResponse struct {
	Error *Error `json:"error"`
	render  render.Render `json:"-"`
}

// RenderError renders a JSON error message with the specified HTTP status code and message
func (l *LogoutResponse) RenderError(code int, message string) {
	// Generate error
	l.Error = new(Error)
	l.Error.Code = code
	l.Error.Message = message

	// Render with specified HTTP status code
	l.render.JSON(code, l)
}

// ServerError is a shortcut to render a HTTP 500 with generic "server error" message
func (l *LogoutResponse) ServerError() {
	l.RenderError(500, "server error")
	return
}

// GetLogout destroys a new session from the wavepipe API, and returns a HTTP status and JSON
func GetLogout(r render.Render, req *http.Request, session *data.Session, params martini.Params) {
	// Output struct for logouts request
	res := LogoutResponse{render: r}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Destroy the current API session
	if err := session.Delete(); err != nil {
		log.Println(err)
		res.ServerError()
		return
	}

	// Build response
	res.Error = nil

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}

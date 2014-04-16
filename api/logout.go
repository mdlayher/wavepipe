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
}

// GetLogout destroys a new session from the wavepipe API, and returns a HTTP status and JSON
func GetLogout(r render.Render, req *http.Request, session *data.Session, params martini.Params) {
	// Output struct for logouts request
	res := LogoutResponse{}

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

	// Destroy the current API session
	if err := session.Delete(); err != nil {
		log.Println(err)

		res.Error = new(Error)
		res.Error.Code = 500
		res.Error.Message = "server error"
		r.JSON(500, res)
		return
	}

	// Build response
	res.Error = nil

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}

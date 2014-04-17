package api

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// GetLogout destroys a new session from the wavepipe API, and returns a HTTP status and JSON
func GetLogout(r render.Render, req *http.Request, session *data.Session, params martini.Params) {
	// Output struct for logout request
	res := ErrorResponse{render: r}

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

package api

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/common"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// StatusResponse represents the JSON response for /api/status
type StatusResponse struct {
	Error  *Error         `json:"error"`
	Status *common.Status `json:"status"`
	render render.Render  `json:"-"`
}

// RenderError renders a JSON error message with the specified HTTP status code and message
func (s *StatusResponse) RenderError(code int, message string) {
	// Nullify all other fields
	s.Status = nil

	// Generate error
	s.Error = new(Error)
	s.Error.Code = code
	s.Error.Message = message

	// Render with specified HTTP status code
	s.render.JSON(code, s)
}

// ServerError is a shortcut to render a HTTP 500 with generic "server error" message
func (s *StatusResponse) ServerError() {
	s.RenderError(500, "server error")
	return
}

// GetStatus returns the current server status, with an HTTP status and JSON
func GetStatus(req *http.Request, r render.Render, params martini.Params) {
	// Output struct for songs request
	res := StatusResponse{render: r}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.RenderError(400, "unsupported API version"+version)
			return
		}
	}

	// Retrieve current server status
	status, err := common.ServerStatus()
	if err != nil {
		log.Println(err)
		res.ServerError()
		return
	}

	// Copy into response
	res.Status = status

	// HTTP 200 OK with JSON
	res.Error = nil
	r.JSON(200, res)
	return
}

package api

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/mdlayher/goset"
)

// APIVersion is the current version of the API
const APIVersion = "v0"

// APIDocumentation provides a link to the current API documentation
const APIDocumentation = "https://github.com/mdlayher/wavepipe/blob/master/API.md"

// apiVersionSet is the set of all currently supported API versions
var apiVersionSet = set.New(APIVersion)

// Error represents an error produced by the API
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse represents the JSON response for endpoints which only return an error
type ErrorResponse struct {
	Error  *Error        `json:"error"`
	render render.Render `json:"-"`
}

// RenderError renders a JSON error message with the specified HTTP status code and message
func (e *ErrorResponse) RenderError(code int, message string) {
	// Render with specified HTTP status code
	e.render.JSON(code, ErrorResponse{
		Error: &Error{
			Code:    code,
			Message: message,
		},
	})
}

// ServerError is a shortcut to render a HTTP 500 with generic "server error" message
func (e *ErrorResponse) ServerError() {
	e.RenderError(500, "server error")
	return
}

// Information represents information about the API
type Information struct {
	Error         *Error   `json:"error"`
	Version       string   `json:"version"`
	Supported     []string `json:"supported"`
	Documentation string   `json:"documentation"`
}

// APIInfo returns information about the API
func APIInfo(r render.Render, params martini.Params) {
	// Output struct for errors
	errRes := ErrorResponse{render: r}

	// Enumerate available API versions
	versions := make([]string, 0)
	for _, v := range apiVersionSet.Enumerate() {
		versions = append(versions, v.(string))
	}

	// Output response
	res := Information{
		Error:         nil,
		Version:       APIVersion,
		Supported:     versions,
		Documentation: APIDocumentation,
	}

	// Check if a "version" was set
	if version, ok := params["version"]; ok {
		// Check if API version is supported
		if !apiVersionSet.Has(version) {
			errRes.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// HTTP 200 OK
	r.JSON(200, res)
}

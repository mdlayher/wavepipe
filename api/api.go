package api

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mdlayher/goset"
	"github.com/unrolled/render"
)

const (
	// Version is the current version of the API
	Version = "v0"
	// Documentation provides a link to the current API documentation
	Documentation = "https://github.com/mdlayher/wavepipe/blob/master/API.md"

	// CtxRender is the key used to store a render instance in gorilla context
	CtxRender = "middleware_render"
	// CtxUser is the key used to store a User instance in gorilla context
	CtxUser = "data_user"
	// CtxSession is the key used to store a Session instance on gorilla context
	CtxSession = "data_session"
)

// apiVersionSet is the set of all currently supported API versions
var apiVersionSet = set.New(Version)

// Error represents an error produced by the API
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse represents the JSON response for endpoints which only return an error
type ErrorResponse struct {
	Error *Error `json:"error"`
}

// errRes generates an ErrorResponse struct containing the specified code and message
func errRes(code int, message string) ErrorResponse {
	return ErrorResponse{
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
}

// permissionErr is the ErrorResponse returned to clients on a permission denied
var permissionErr = errRes(403, "permission denied")

// serverErr is the ErrorResponse returned to clients on an internal server error
var serverErr = errRes(500, "server error")

// Information represents information about the API
type Information struct {
	Error         *Error   `json:"error"`
	Version       string   `json:"version"`
	Supported     []string `json:"supported"`
	Documentation string   `json:"documentation"`
}

// APIInfo returns information about the API
func APIInfo(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Enumerate available API versions
	versions := make([]string, 0)
	for _, v := range apiVersionSet.Enumerate() {
		versions = append(versions, v.(string))
	}

	// Output response
	info := Information{
		Error:         nil,
		Version:       Version,
		Supported:     versions,
		Documentation: Documentation,
	}

	// Check if a "version" was set
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if API version is supported
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// HTTP 200 OK
	r.JSON(res, 200, info)
}

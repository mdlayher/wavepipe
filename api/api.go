package api

import (
	"net/http"

	"github.com/martini-contrib/render"
)

// APIVersion is the current version of the API
const APIVersion = "v0"

// APIDocumentation provides a link to the current API documentation
const APIDocumentation = "https://github.com/mdlayher/wavepipe/blob/master/API.md"

// endpoints is the current list of supported API endpoints
var endpoints []string = []string{
	"/api/albums",
	"/api/artists",
	"/api/songs",
}

// Error represents an error produced by the API
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Information represents information about the API
type Information struct {
	Error         *Error   `json:"error"`
	Version       string   `json:"version"`
	Documentation string   `json:"documentation"`
	Endpoints     []string `json:"endpoints"`
}

// APIInfo returns information about the API
func APIInfo(r render.Render) {
	// HTTP 200 OK with JSON
	r.JSON(http.StatusOK, Information{
		Error:         nil,
		Version:       APIVersion,
		Documentation: APIDocumentation,
		Endpoints:     endpoints,
	})
}

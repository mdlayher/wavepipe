package api

import (
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/mdlayher/goset"
)

// APIVersion is the current version of the API
const APIVersion = "v0"

// APIDocumentation provides a link to the current API documentation
const APIDocumentation = "https://github.com/mdlayher/wavepipe/blob/master/API.md"

// apiVersionSet is the set of all currently supported API versions
var apiVersionSet = set.New("v0")

// endpoints is a list of supported API endpoints
var endpoints = []string{
	"/api/v0/albums",
	"/api/v0/artists",
	"/api/v0/songs",
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
	Supported     []string `json:"supported"`
	Documentation string   `json:"documentation"`
	Endpoints     []string `json:"endpoints"`
}

// APIInfo returns information about the API
func APIInfo(r render.Render, params martini.Params) {
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
		Endpoints:     endpoints,
	}

	// Check if a "version" was set
	if version, ok := params["version"]; ok {
		// Check if API version is supported
		if !apiVersionSet.Has(version) {
			res.Error = new(Error)
			res.Error.Code = http.StatusBadRequest
			res.Error.Message = "unsupported API version: " + version
			r.JSON(http.StatusBadRequest, res)
			return
		}
	}

	// HTTP 200 OK
	r.JSON(http.StatusOK, res)
}

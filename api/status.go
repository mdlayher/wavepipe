package api

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/common"

	"github.com/go-martini/martini"
	"github.com/mdlayher/render"
)

// StatusResponse represents the JSON response for /api/status
type StatusResponse struct {
	Error  *Error         `json:"error"`
	Status *common.Status `json:"status"`
}

// GetStatus returns the current server status, with an HTTP status and JSON
func GetStatus(req *http.Request, r render.Render, params martini.Params) {
	// Output struct for songs request
	res := StatusResponse{}

	// Output struct for errors
	errRes := ErrorResponse{render: r}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			errRes.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Retrieve current server status
	status, err := common.ServerStatus()
	if err != nil {
		log.Println(err)
		errRes.ServerError()
		return
	}

	// Copy into response
	res.Status = status

	// HTTP 200 OK with JSON
	res.Error = nil
	r.JSON(200, res)
	return
}

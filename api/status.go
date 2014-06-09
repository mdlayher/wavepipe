package api

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/common"
	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// StatusResponse represents the JSON response for /api/status
type StatusResponse struct {
	Error   *Error         `json:"error"`
	Status  *common.Status `json:"status"`
	Metrics *Metrics       `json:"metrics"`
}

// Metrics represents a variety of metrics about wavepipe's database
type Metrics struct {
	Artists int64 `json:"artists"`
	Albums  int64 `json:"albums"`
	Songs   int64 `json:"songs"`
	Folders int64 `json:"folders"`
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

	// If requested, fetch additional metrics (not added by default due to full table scans in database)
	if pMetrics := req.URL.Query().Get("metrics"); pMetrics == "true" {
		// Fetch total artists
		artists, err := data.DB.CountArtists()
		if err != nil {
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Fetch total albums
		albums, err := data.DB.CountAlbums()
		if err != nil {
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Fetch total songs
		songs, err := data.DB.CountSongs()
		if err != nil {
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Fetch total folders
		folders, err := data.DB.CountFolders()
		if err != nil {
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Add metrics to output
		res.Metrics = &Metrics{
			Artists: artists,
			Albums:  albums,
			Songs:   songs,
			Folders: folders,
		}
	}

	// HTTP 200 OK with JSON
	res.Error = nil
	r.JSON(200, res)
	return
}

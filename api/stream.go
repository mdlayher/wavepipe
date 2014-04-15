package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// StreamResponse represents the JSON response for /api/streams
type StreamResponse struct {
	Error *Error `json:"error"`
}

// GetStream a raw, non-transcoded, media file stream from wavepipe.  On success, this API will
// return a binary stream. On failure, it will return a JSON error.
func GetStream(r render.Render, params martini.Params) {
	// Output struct for streams request
	res := StreamResponse{}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.Error = new(Error)
			res.Error.Code = http.StatusBadRequest
			res.Error.Message = "unsupported API version: " + version
			r.JSON(http.StatusBadRequest, res)
			return
		}
	}

	// Check for an ID parameter
	pID, ok := params["id"]
	if !ok {
		res.Error = new(Error)
		res.Error.Code = http.StatusBadRequest
		res.Error.Message = "no integer stream ID provided"
		r.JSON(http.StatusBadRequest, res)
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		res.Error = new(Error)
		res.Error.Code = http.StatusBadRequest
		res.Error.Message = "invalid integer stream ID"
		r.JSON(http.StatusBadRequest, res)
		return
	}

	// Attempt to load the song with matching ID
	song := new(data.Song)
	song.ID = id
	if err := song.Load(); err != nil {
		res.Error = new(Error)

		// Check for invalid ID
		if err == sql.ErrNoRows {
			res.Error.Code = http.StatusNotFound
			res.Error.Message = "song ID not found"
			r.JSON(http.StatusNotFound, res)
			return
		}

		// All other errors
		log.Println(err)
		res.Error.Code = http.StatusInternalServerError
		res.Error.Message = "server error"
		r.JSON(http.StatusInternalServerError, res)
		return
	}

	// Attempt to access data stream
	// TODO: implement method to open Song's stream

	// Build response
	res.Error = nil

	// HTTP 200 OK with JSON
	r.JSON(http.StatusOK, res)
	return
}

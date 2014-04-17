package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

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
func GetStream(httpRes http.ResponseWriter, r render.Render, params martini.Params) {
	// Output struct for streams request
	res := StreamResponse{}

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

	// Check for an ID parameter
	pID, ok := params["id"]
	if !ok {
		res.Error = new(Error)
		res.Error.Code = 400
		res.Error.Message = "no integer stream ID provided"
		r.JSON(400, res)
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		res.Error = new(Error)
		res.Error.Code = 400
		res.Error.Message = "invalid integer stream ID"
		r.JSON(400, res)
		return
	}

	// Attempt to load the song with matching ID
	song := new(data.Song)
	song.ID = id
	if err := song.Load(); err != nil {
		res.Error = new(Error)

		// Check for invalid ID
		if err == sql.ErrNoRows {
			res.Error.Code = 404
			res.Error.Message = "song ID not found"
			r.JSON(404, res)
			return
		}

		// All other errors
		log.Println(err)
		res.Error.Code = 500
		res.Error.Message = "server error"
		r.JSON(500, res)
		return
	}

	// Attempt to access data stream
	stream, err := song.Stream()
	if err != nil {
		log.Println(err)

		res.Error = new(Error)
		res.Error.Code = 500
		res.Error.Message = "server error"
		r.JSON(500, res)
		return
	}
	defer stream.Close()

	// Attempt to send file stream over HTTP
	log.Printf("stream: starting: [#%05d] %s - %s ", song.ID, song.Artist, song.Title)
	if err := httpStream(song, song.FileSize, stream, httpRes); err != nil {
		// Check for client reset
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
			return
		}

		log.Println("stream: error:", err)
		return
	}

	log.Printf("stream: completed: [#%05d] %s - %s", song.ID, song.Artist, song.Title)
	return
}

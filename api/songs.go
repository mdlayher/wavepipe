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

// SongsResponse represents the JSON response for /api/songs
type SongsResponse struct {
	Error *Error      `json:"error"`
	Songs []data.Song `json:"songs"`
}

// GetSongs retrieves one or more songs from wavepipe, and returns a HTTP status and JSON
func GetSongs(r render.Render, params martini.Params) {
	// Output struct for songs request
	res := SongsResponse{}

	// List of songs to return
	songs := make([]data.Song, 0)

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
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.Error = new(Error)
			res.Error.Code = 400
			res.Error.Message = "invalid integer song ID"
			r.JSON(400, res)
			return
		}

		// Load the song
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
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// Add song to slice
		songs = append(songs, *song)
	} else {
		// Retrieve all songs
		tempSongs, err := data.DB.AllSongs()
		if err != nil {
			log.Println(err)
			res.Error = new(Error)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// Copy songs into the output slice
		songs = tempSongs
	}

	// Build response
	res.Error = nil
	res.Songs = songs

	// HTTP 200 OK with JSON
	r.JSON(http.StatusOK, res)
	return
}

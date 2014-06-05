package api

import (
	"database/sql"
	"fmt"
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
func GetSongs(r render.Render, req *http.Request, params martini.Params) {
	// Output struct for songs request
	res := SongsResponse{}

	// Output struct for errors
	errRes := ErrorResponse{render: r}

	// List of songs to return
	songs := make([]data.Song, 0)

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			errRes.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			errRes.RenderError(400, "invalid integer song ID")
			return
		}

		// Load the song
		song := new(data.Song)
		song.ID = id
		if err := song.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				errRes.RenderError(404, "song ID not found")
				return
			}

			// All other errors
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Add song to slice
		songs = append(songs, *song)
	} else {
		// Check for a limit parameter
		if pLimit := req.URL.Query().Get("limit"); pLimit != "" {
			// Split limit into two integers
			var offset int
			var count int
			if n, err := fmt.Sscanf(pLimit, "%d,%d", &offset, &count); n < 2 || err != nil {
				errRes.RenderError(400, "invalid comma-separated integer pair for limit")
				return
			}

			// Retrieve limited subset of songs
			tempSongs, err := data.DB.LimitSongs(offset, count)
			if err != nil {
				log.Println(err)
				errRes.ServerError()
				return
			}

			// Copy songs into the output slice
			songs = tempSongs
		} else {
			// Retrieve all songs
			tempSongs, err := data.DB.AllSongs()
			if err != nil {
				log.Println(err)
				errRes.ServerError()
				return
			}

			// Copy songs into the output slice
			songs = tempSongs
		}
	}

	// Build response
	res.Error = nil
	res.Songs = songs

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}

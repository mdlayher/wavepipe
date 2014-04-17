package api

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// SongsResponse represents the JSON response for /api/songs
type SongsResponse struct {
	Error  *Error        `json:"error"`
	Songs  []data.Song   `json:"songs"`
	render render.Render `json:"-"`
}

// RenderError renders a JSON error message with the specified HTTP status code and message
func (s *SongsResponse) RenderError(code int, message string) {
	// Nullify all other fields
	s.Songs = nil

	// Generate error
	s.Error = new(Error)
	s.Error.Code = code
	s.Error.Message = message

	// Render with specified HTTP status code
	s.render.JSON(code, s)
}

// ServerError is a shortcut to render a HTTP 500 with generic "server error" message
func (a *SongsResponse) ServerError() {
	a.RenderError(500, "server error")
	return
}

// GetSongs retrieves one or more songs from wavepipe, and returns a HTTP status and JSON
func GetSongs(r render.Render, params martini.Params) {
	// Output struct for songs request
	res := SongsResponse{render: r}

	// List of songs to return
	songs := make([]data.Song, 0)

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.RenderError(400, "unsupported API version"+version)
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.RenderError(400, "invalid integer song ID")
			return
		}

		// Load the song
		song := new(data.Song)
		song.ID = id
		if err := song.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				res.RenderError(404, "song ID not found")
				return
			}

			// All other errors
			log.Println(err)
			res.ServerError()
			return
		}

		// Add song to slice
		songs = append(songs, *song)
	} else {
		// Retrieve all songs
		tempSongs, err := data.DB.AllSongs()
		if err != nil {
			log.Println(err)
			res.ServerError()
			return
		}

		// Copy songs into the output slice
		songs = tempSongs
	}

	// Build response
	res.Error = nil
	res.Songs = songs

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}

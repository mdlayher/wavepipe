package api

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// AlbumsResponse represents the JSON response for /api/albums
type AlbumsResponse struct {
	Error  *Error        `json:"error"`
	Albums []data.Album  `json:"albums"`
	Songs  []data.Song   `json:"songs"`
	render render.Render `json:"-"`
}

// RenderError renders a JSON error message with the specified HTTP status code and message
func (a *AlbumsResponse) RenderError(code int, message string) {
	// Nullify all other fields
	a.Albums = nil
	a.Songs = nil

	// Generate error
	a.Error = new(Error)
	a.Error.Code = code
	a.Error.Message = message

	// Render with specified HTTP status code
	a.render.JSON(code, a)
}

// ServerError is a shortcut to render a HTTP 500 with generic "server error" message
func (a *AlbumsResponse) ServerError() {
	a.RenderError(500, "server error")
	return
}

// GetAlbums retrieves one or more albums from wavepipe, and returns a HTTP status and JSON
func GetAlbums(r render.Render, params martini.Params) {
	// Output struct for albums request
	res := AlbumsResponse{render: r}

	// List of albums to return
	albums := make([]data.Album, 0)

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.RenderError(400, "invalid integer album ID")
			return
		}

		// Load the album
		album := new(data.Album)
		album.ID = id
		if err := album.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				res.RenderError(404, "album ID not found")
				return
			}

			// All other errors
			log.Println(err)
			res.ServerError()
			return
		}

		// On single album, load the songs for this album
		songs, err := data.DB.SongsForAlbum(album.ID)
		if err != nil {
			log.Println(err)
			res.ServerError()
			return
		}

		// Add songs to output
		res.Songs = songs

		// Add album to slice
		albums = append(albums, *album)
	} else {
		// Retrieve all albums
		tempAlbums, err := data.DB.AllAlbums()
		if err != nil {
			log.Println(err)
			res.ServerError()
			return
		}

		// Copy albums into the output slice
		albums = tempAlbums
	}

	// Build response
	res.Error = nil
	res.Albums = albums

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}

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

// AlbumsResponse represents the JSON response for /api/albums
type AlbumsResponse struct {
	Error  *Error       `json:"error"`
	Albums []data.Album `json:"albums"`
	Songs  []data.Song  `json:"songs"`
}

// GetAlbums retrieves one or more albums from wavepipe, and returns a HTTP status and JSON
func GetAlbums(r render.Render, params martini.Params) {
	// Output struct for albums request
	res := AlbumsResponse{}

	// List of albums to return
	albums := make([]data.Album, 0)

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.Error = new(Error)
			res.Error.Message = "invalid integer album ID"
			r.JSON(http.StatusBadRequest, res)
			return
		}

		// Load the album
		album := new(data.Album)
		album.ID = id
		if err := album.Load(); err != nil {
			res.Error = new(Error)

			// Check for invalid ID
			if err == sql.ErrNoRows {
				res.Error.Message = "album ID not found"
				r.JSON(http.StatusNotFound, res)
				return
			}

			// All other errors
			log.Println(err)
			res.Error.Message = "server error"
			r.JSON(http.StatusInternalServerError, res)
			return
		}

		// On single album, load the songs for this album
		songs, err := data.DB.SongsForAlbum(album.ID)
		if err != nil {
			log.Println(err)
			res.Error = new(Error)
			res.Error.Message = "server error"
			r.JSON(http.StatusInternalServerError, res)
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
			res.Error = new(Error)
			res.Error.Message = "server error"
			r.JSON(http.StatusInternalServerError, res)
			return
		}

		// Copy albums into the output slice
		albums = tempAlbums
	}

	// Build response
	res.Error = nil
	res.Albums = albums

	// HTTP 200 OK with JSON
	r.JSON(http.StatusOK, res)
	return
}
